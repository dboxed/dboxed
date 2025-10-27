package webdavproxy

import (
	"io/fs"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/dboxed/dboxed/pkg/server/models"
	"golang.org/x/net/webdav"
)

type dir struct {
	fs *FileSystem

	prefix string

	m        sync.Mutex
	lastUsed time.Time
	fis      []fs.FileInfo
}

func (d *dir) Stat() (fs.FileInfo, error) {
	return &dirInfo{
		name: d.prefix,
	}, nil
}

func (d *dir) Close() error {
	return nil
}

func (d *dir) Read(p []byte) (n int, err error) {
	return 0, webdav.ErrNotImplemented
}

func (d *dir) Seek(offset int64, whence int) (int64, error) {
	return 0, webdav.ErrNotImplemented
}

func (d *dir) Readdir(count int) ([]fs.FileInfo, error) {
	d.m.Lock()
	defer d.m.Unlock()

	d.lastUsed = time.Now()

	if d.fis != nil {
		return d.fis, nil
	}

	rep, err := d.fs.client2.S3ProxyListObjects(d.fs.ctx, d.fs.s3BucketId, models.S3ProxyListObjectsRequest{
		Prefix: path.Join(d.fs.s3Prefix, d.prefix) + "/",
	})
	if err != nil {
		return nil, err
	}

	var ret []fs.FileInfo
	for _, x := range rep.Objects {
		key := strings.TrimPrefix(x.Key, d.fs.s3Prefix)
		key = strings.TrimPrefix(key, "/")

		if strings.HasSuffix(key, "/") {
			ret = append(ret, &dirInfo{
				name: strings.TrimSuffix(key, "/"),
			})
		} else {
			ret = append(ret, &fileInfo{
				oi: x,
			})
		}
	}

	d.fis = ret

	if count != 0 && len(ret) > count {
		ret = ret[:count]
	}
	return ret, nil
}

func (d *dir) Write(p []byte) (n int, err error) {
	return 0, webdav.ErrNotImplemented
}
