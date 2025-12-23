package download_helper

import (
	"compress/bzip2"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/util"
	"github.com/klauspost/compress/gzip"
	"github.com/rogpeppe/go-internal/lockedfile"
)

type Opts struct {
	Url         string
	DownloadKey string
	BinName     string

	Gz  bool
	Bz2 bool
}

func DownloadBinary(ctx context.Context, opts Opts) (string, error) {
	dir := filepath.Join(util.GetUserCacheDir(ctx), "download-binaries", opts.DownloadKey)
	file := filepath.Join(dir, opts.BinName)

	if _, err := os.Stat(file); err == nil {
		return dir, nil
	}

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}

	lockPath := filepath.Join(dir, ".download-lock")
	lock, err := lockedfile.Create(lockPath)
	if err != nil {
		return "", err
	}
	defer lock.Close()

	slog.InfoContext(ctx, "downloading binary", "url", opts.Url, "downloadKey", opts.DownloadKey, "binName", opts.BinName)

	req, err := http.NewRequest("GET", opts.Url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("downloading binary from %s failed: %s", opts.Url, resp.Status)
	}

	var s io.Reader = resp.Body
	if opts.Gz {
		x, err := gzip.NewReader(s)
		if err != nil {
			return "", err
		}
		defer x.Close()
		s = x
	}
	if opts.Bz2 {
		s = bzip2.NewReader(s)
	}

	tmpFile, err := os.CreateTemp(dir, ".download-")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	_, err = io.Copy(tmpFile, s)
	if err != nil {
		return "", fmt.Errorf("error while reading download stream for %s: %w", opts.Url, err)
	}
	err = tmpFile.Close()
	if err != nil {
		return "", err
	}

	err = os.Chmod(tmpFile.Name(), 0755)
	if err != nil {
		return "", fmt.Errorf("error while performing chmod for %s: %w", opts.Url, err)
	}

	err = os.Rename(tmpFile.Name(), file)
	if err != nil {
		return "", fmt.Errorf("error while performing final rename for %s: %w", opts.Url, err)
	}
	return dir, nil
}
