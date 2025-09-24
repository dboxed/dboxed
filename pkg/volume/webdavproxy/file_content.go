package webdavproxy

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dustin/go-humanize"
)

const chunkSize = 4096
const superChunk = humanize.MiByte

const maxCachedBytes = humanize.MiByte * 2
const maxCachedChunks = maxCachedBytes / chunkSize
const maxCachedChunksCleanLimit = maxCachedChunks * 2

type fileContent struct {
	fs *FileSystem
	oi models.S3ObjectInfo

	m            sync.Mutex
	chunks       []fileChunk
	filledChunks int

	lastFillEnd int
	cnt         int

	lastUsed time.Time
}

type fileChunk struct {
	buf []byte

	lastReadTime time.Time
}

func newFileContent(fs *FileSystem, oi models.S3ObjectInfo) *fileContent {
	f := &fileContent{
		fs:       fs,
		oi:       oi,
		chunks:   make([]fileChunk, (oi.Size+chunkSize-1)/chunkSize),
		lastUsed: time.Now(),
	}
	return f
}

func (f *fileContent) forgetOldChunks(lock bool) {
	if lock {
		f.m.Lock()
		defer f.m.Unlock()
	}
	if f.filledChunks < maxCachedChunksCleanLimit {
		return
	}

	sorted := make([]*fileChunk, 0, len(f.chunks))
	for i := range f.chunks {
		if f.chunks[i].buf == nil {
			continue
		}
		sorted = append(sorted, &f.chunks[i])
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].lastReadTime.After(sorted[j].lastReadTime)
	})

	for i := len(sorted) - 1; i > maxCachedChunks; i-- {
		sorted[i].buf = nil
		f.filledChunks--
	}
}

func (f *fileContent) fillChunks(firstChunk int, lastChunk int) error {
	firstByte := firstChunk * chunkSize
	lastByte := lastChunk*chunkSize + chunkSize - 1

	rangeHeader := fmt.Sprintf("bytes=%d-%d", firstByte, lastByte)

	slog.Info("fillChunks", slog.Any("key", f.oi.Key),
		slog.Any("chunks", fmt.Sprintf("%d-%d", firstChunk, lastChunk)),
		slog.Any("bytes", fmt.Sprintf("%d-%d (%s)", firstByte, lastByte, humanize.Bytes(uint64(lastByte-firstByte+1)))))

	req, err := http.NewRequest("GET", f.oi.PresignedGetUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Range", rangeHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == 404 {
			return os.ErrNotExist
		}
		return fmt.Errorf("http status: %s", resp.Status)
	}

	for i := firstChunk; i <= lastChunk; i++ {
		buf := make([]byte, chunkSize)
		n, err := io.ReadFull(resp.Body, buf)
		if err != nil {
			if err != io.EOF && !errors.Is(err, io.ErrUnexpectedEOF) {
				return err
			}
		}
		c := fileChunk{
			buf:          buf[:n],
			lastReadTime: time.Now(),
		}
		f.chunks[i] = c
	}

	f.filledChunks += lastChunk - firstChunk + 1

	return nil
}

func (f *fileContent) ensureChunks(offset int64, count int) error {
	firstChunk := int(offset / chunkSize)
	lastChunk := int((offset + int64(count)) / chunkSize)

	for firstChunk < len(f.chunks) && f.chunks[firstChunk].buf != nil {
		firstChunk++
	}
	if firstChunk > lastChunk {
		return nil
	}
	for lastChunk >= 0 && f.chunks[lastChunk].buf != nil {
		lastChunk--
	}
	if lastChunk < firstChunk {
		return nil
	}
	if firstChunk == lastChunk && f.chunks[firstChunk].buf != nil {
		return nil
	}

	if f.lastFillEnd == firstChunk-1 {
		f.cnt++
	} else {
		f.cnt = 0
	}

	if f.cnt >= 4 {
		for lastChunk-firstChunk+1 < superChunk/chunkSize && lastChunk+1 < len(f.chunks) && f.chunks[lastChunk+1].buf == nil {
			lastChunk++
		}
	}
	f.lastFillEnd = lastChunk

	return f.fillChunks(firstChunk, lastChunk)
}

func (f *fileContent) readAt(p []byte, offset int64) (int, error) {
	f.m.Lock()
	defer f.m.Unlock()

	f.lastUsed = time.Now()

	err := f.ensureChunks(offset, len(p))
	if err != nil {
		return 0, err
	}

	pos := 0
	for pos < len(p) {
		i := offset / chunkSize
		chunkOffset := offset % chunkSize

		n := copy(p[pos:], f.chunks[i].buf[chunkOffset:])
		pos += n
		offset += int64(n)

		f.chunks[i].lastReadTime = time.Now()
	}

	f.forgetOldChunks(false)

	return pos, nil
}
