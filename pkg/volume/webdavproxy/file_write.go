package webdavproxy

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"sync"

	"github.com/dboxed/dboxed/pkg/server/models"
	"golang.org/x/net/webdav"
)

type fileWrite struct {
	fs  *FileSystem
	key string

	presignedPut *models.S3ProxyPresignPutResult

	m            sync.Mutex
	uploadReq    *http.Request
	uploadResp   *http.Response
	uploadWriter *io.PipeWriter
	uploadErr    error
	uploadDone   chan struct{}
	written      int64
}

func (f *fileWrite) Stat() (fs.FileInfo, error) {
	return &fileInfo{
		oi: models.S3ObjectInfo{
			Key:  f.key,
			Size: f.written,
		},
	}, nil
}

func (f *fileWrite) presignPutUrl() error {
	slog.Info("presignPutUrl", slog.Any("key", f.key))

	rep, err := f.fs.client2.S3ProxyPresignPut(f.fs.ctx, f.fs.volumeProviderId, models.S3ProxyPresignPutRequest{
		Key: f.key,
	})
	if err != nil {
		return err
	}
	f.presignedPut = rep
	return nil
}

func (f *fileWrite) beginUpload(url string) error {
	bodyReader, bodyWriter := io.Pipe()

	f.m.Lock()
	defer f.m.Unlock()

	if f.uploadReq != nil {
		return fmt.Errorf("upload already started")
	}

	slog.Info("beginUpload", slog.Any("key", f.key))

	req, err := http.NewRequest("PUT", url, bufio.NewReaderSize(bodyReader, 1024*64))
	if err != nil {
		return err
	}

	f.uploadReq = req
	f.uploadDone = make(chan struct{})
	f.uploadWriter = bodyWriter
	go func() {
		resp, err := http.DefaultClient.Do(req)
		defer resp.Body.Close()
		f.handleUploadDone(resp, err)
	}()

	return nil
}

func (f *fileWrite) handleUploadDone(resp *http.Response, err error) {
	f.m.Lock()
	defer f.m.Unlock()

	slog.Info("handleUploadDone", slog.Any("key", f.key))

	f.uploadResp = resp
	f.uploadErr = err
	close(f.uploadDone)
}

func (f *fileWrite) checkUploadErr() error {
	f.m.Lock()
	defer f.m.Unlock()
	if f.uploadErr != nil {
		return f.uploadErr
	}
	if f.uploadResp == nil {
		return nil
	}
	if f.uploadResp.StatusCode < 200 || f.uploadResp.StatusCode >= 300 {
		return fmt.Errorf("upload status: %s", f.uploadResp.Status)
	}
	return nil
}

func (f *fileWrite) Start() error {
	err := f.presignPutUrl()
	if err != nil {
		return err
	}

	err = f.beginUpload(f.presignedPut.PresignedUrl)
	if err != nil {
		return err
	}
	return nil
}

func (f *fileWrite) Close() error {
	defer func() {
		f.fs.forgetCache(f.key, true)
	}()

	err := f.checkUploadErr()
	if err != nil {
		return err
	}

	//slog.Info("Close", slog.Any("key", f.key))

	err = f.uploadWriter.Close()
	if err != nil {
		f.m.Lock()
		if f.uploadErr != nil {
			f.uploadErr = err
		}
		f.m.Unlock()
		return err
	}

	//slog.Info("wait for close", slog.Any("key", f.key))
	<-f.uploadDone
	//slog.Info("close done", slog.Any("key", f.key))

	err = f.checkUploadErr()
	if err != nil {
		return err
	}

	return nil
}

func (f *fileWrite) Write(p []byte) (int, error) {
	//slog.Info("Write", slog.Any("key", f.key), slog.Any("len", len(p)))
	return f.uploadWriter.Write(p)
}

func (f *fileWrite) Seek(offset int64, whence int) (int64, error) {
	return 0, webdav.ErrNotImplemented
}

func (f *fileWrite) Read(p []byte) (n int, err error) {
	return 0, webdav.ErrNotImplemented
}

func (f *fileWrite) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, webdav.ErrNotImplemented
}
