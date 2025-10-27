package webdavproxy

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/dboxed/dboxed/pkg/server/models"
	"golang.org/x/net/webdav"
)

type fileWrite struct {
	fs  *FileSystem
	key string

	presignedPut *models.S3ProxyPresignPutResult

	tmpFile *os.File
}

func (f *fileWrite) Stat() (fs.FileInfo, error) {
	st, err := f.tmpFile.Stat()
	if err != nil {
		return nil, err
	}
	return &fileInfo{
		oi: models.S3ObjectInfo{
			Key:  f.key,
			Size: st.Size(),
		},
	}, nil
}

func (f *fileWrite) presignPutUrl() error {
	slog.Debug("presignPutUrl", slog.Any("key", f.key))

	key := path.Join(f.fs.s3Prefix, f.key)
	key = strings.TrimPrefix(key, "/")

	rep, err := f.fs.client2.S3ProxyPresignPut(f.fs.ctx, f.fs.s3BucketId, models.S3ProxyPresignPutRequest{
		Key: key,
	})
	if err != nil {
		return err
	}
	f.presignedPut = rep
	return nil
}

func (f *fileWrite) upload() error {
	_, err := f.tmpFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	st, err := f.tmpFile.Stat()
	if err != nil {
		return err
	}

	slog.Debug("upload", slog.Any("key", f.key), slog.Any("size", st.Size()))

	req, err := http.NewRequest("PUT", f.presignedPut.PresignedUrl, f.tmpFile)
	if err != nil {
		return err
	}
	req.ContentLength = st.Size()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("upload request returned status %s", resp.Status)
	}

	return nil
}

func (f *fileWrite) Start() error {
	err := f.presignPutUrl()
	if err != nil {
		return err
	}

	f.tmpFile, err = os.CreateTemp("", "")
	if err != nil {
		return err
	}
	// we remove the file immediately and keep it open, so we can be sure it's deleted when the
	// process exits
	err = os.Remove(f.tmpFile.Name())
	if err != nil {
		return err
	}

	return nil
}

func (f *fileWrite) Close() error {
	defer func() {
		f.fs.forgetCache(f.key, true)
		_ = f.tmpFile.Close()
	}()

	slog.Debug("close", slog.Any("key", f.key))

	err := f.upload()
	if err != nil {
		return err
	}

	return nil
}

func (f *fileWrite) Write(p []byte) (int, error) {
	slog.Debug("write", slog.Any("key", f.key), slog.Any("len", len(p)))
	return f.tmpFile.Write(p)
}

func (f *fileWrite) Seek(offset int64, whence int) (int64, error) {
	return f.tmpFile.Seek(offset, whence)
}

func (f *fileWrite) Read(p []byte) (n int, err error) {
	return f.tmpFile.Read(p)
}

func (f *fileWrite) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, webdav.ErrNotImplemented
}
