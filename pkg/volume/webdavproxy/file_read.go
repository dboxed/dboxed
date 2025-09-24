package webdavproxy

import (
	"errors"
	"io"
	"io/fs"
	"time"

	"golang.org/x/net/webdav"
)

type fileRead struct {
	content *fileContent

	readOffset int64
}

func (f *fileRead) Stat() (fs.FileInfo, error) {
	return &fileInfo{
		oi: f.content.oi,
	}, nil
}

func (f *fileRead) Close() error {
	//slog.Info("close", slog.Any("key", f.content.statResult.ObjectInfo.Key))
	return nil
}

func (f *fileRead) Read(p []byte) (int, error) {
	f.content.lastUsed = time.Now()
	if len(p) == 0 {
		return 0, nil
	}

	n, err := f.content.readAt(p, f.readOffset)
	f.readOffset += int64(n)
	if err != nil {
		return n, err
	}
	return n, nil
}

func (f *fileRead) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = f.readOffset + offset
	case io.SeekEnd:
		abs = f.content.oi.Size + offset
	default:
		return 0, errors.New("invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("negative position")
	}

	f.readOffset = abs
	return abs, nil
}

func (f *fileRead) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, webdav.ErrNotImplemented
}

func (f *fileRead) Write(p []byte) (n int, err error) {
	return 0, webdav.ErrNotImplemented
}
