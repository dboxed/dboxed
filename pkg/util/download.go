package util

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type CompressionType int

const (
	CompressionNone = iota
	CompressionGzip
	CompressionZstd
)

type readCloserWrapper struct {
	r io.ReadCloser
	c io.ReadCloser
}

func (r *readCloserWrapper) Read(p []byte) (n int, err error) {
	return r.r.Read(p)
}

func (r *readCloserWrapper) Close() error {
	err1 := r.c.Close()
	err2 := r.r.Close()
	if err1 != nil || err2 != nil {
		if err2 != nil {
			return err2
		}
		return err1
	}
	return nil
}

func DownloadStream(ctx context.Context, url string, compression CompressionType) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	closeBody := true
	defer func() {
		if closeBody {
			resp.Body.Close()
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("http status %d", resp.StatusCode)
	}

	contentLengthStr := resp.Header.Get("Content-Length")
	contentLength, _ := strconv.ParseInt(contentLengthStr, 10, 64)

	slog.InfoContext(ctx, "download size determined", slog.Any("downloadSize", humanize.IBytes(uint64(contentLength))))

	r := resp.Body
	switch compression {
	case CompressionNone:
	case CompressionGzip:
		c, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}
		r = &readCloserWrapper{
			r: r,
			c: c,
		}
	case CompressionZstd:
		c, err := zstd.NewReader(r)
		if err != nil {
			return nil, err
		}
		r = &readCloserWrapper{
			r: r,
			c: c.IOReadCloser(),
		}
	default:
		return nil, fmt.Errorf("compression not implemented")
	}

	closeBody = false
	return r, nil
}

func DownloadFile(ctx context.Context, url string, hash string, dir string, compression CompressionType) (string, error) {
	if hash != "" {
		pth := filepath.Join(dir, hash)
		fileHash, err := Sha256SumFile(pth)
		if err != nil {
			if !os.IsNotExist(err) {
				return "", err
			}
		}
		if fileHash == hash {
			return pth, nil
		}
	}

	s, err := DownloadStream(ctx, url, compression)
	if err != nil {
		return "", err
	}
	defer s.Close()

	tmpFile, err := os.CreateTemp(dir, "")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	h := sha256.New()
	r := io.TeeReader(s, h)

	_, err = io.Copy(tmpFile, r)
	if err != nil {
		return "", err
	}
	err = tmpFile.Close()
	if err != nil {
		return "", err
	}

	downloadHash := hex.EncodeToString(h.Sum(nil))
	if hash != "" && downloadHash != hash {
		return "", fmt.Errorf("unexpeted download hash: got %s, expected %s", downloadHash, hash)
	}

	pth := filepath.Join(dir, downloadHash)
	err = os.Rename(tmpFile.Name(), pth)
	if err != nil {
		return "", err
	}

	return pth, nil
}
