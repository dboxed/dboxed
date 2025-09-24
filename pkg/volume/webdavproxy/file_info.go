package webdavproxy

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dboxed/dboxed/pkg/server/models"
)

type fileInfo struct {
	oi models.S3ObjectInfo
}

func (f *fileInfo) Name() string {
	return filepath.Base(f.oi.Key)
}
func (f *fileInfo) Size() int64 {
	return f.oi.Size
}
func (f *fileInfo) Mode() os.FileMode {
	return 0
}

func (f *fileInfo) ModTime() time.Time {
	if f.oi.LastModified != nil {
		return *f.oi.LastModified
	} else {
		return time.Time{}
	}
}
func (f *fileInfo) IsDir() bool      { return false }
func (f *fileInfo) Sys() interface{} { return nil }

func (f *fileInfo) ContentType(ctx context.Context) (string, error) {
	return "", nil
}

func (f *fileInfo) ETag(ctx context.Context) (string, error) {
	return f.oi.Etag, nil
}

type dirInfo struct {
	name string
}

func (f *dirInfo) Name() string {
	return filepath.Base(strings.TrimSuffix(f.name, "/"))
}
func (f *dirInfo) Size() int64 {
	return 0
}
func (f *dirInfo) Mode() os.FileMode {
	return os.ModeDir
}

func (f *dirInfo) ModTime() time.Time {
	return time.Time{}

}
func (f *dirInfo) IsDir() bool      { return true }
func (f *dirInfo) Sys() interface{} { return nil }
