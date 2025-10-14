package webdavproxy

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
	"golang.org/x/net/webdav"
)

type FileSystem struct {
	ctx              context.Context
	client           *baseclient.Client
	client2          *clients.S3ProxyClient
	volumeProviderId int64

	m            sync.Mutex
	dirCache     map[string]*dir
	contentCache map[string]*fileContent
}

func NewFileSystem(ctx context.Context, client *baseclient.Client, volumeProviderId int64) *FileSystem {
	return &FileSystem{
		ctx:              ctx,
		client:           client,
		client2:          &clients.S3ProxyClient{Client: client},
		volumeProviderId: volumeProviderId,

		dirCache:     map[string]*dir{},
		contentCache: map[string]*fileContent{},
	}
}

func (fs *FileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return nil
}

func (fs *FileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	if flag != 0 {
		return fs.doOpenFile(ctx, name, flag, perm)
	}

	st, err := fs.Stat(ctx, name)
	if err != nil {
		return nil, err
	}

	key := normalizeName(name)

	fs.m.Lock()
	fs.cleanupCache()
	fs.m.Unlock()

	if st.IsDir() {
		fs.m.Lock()
		dc, ok := fs.dirCache[key]
		fs.m.Unlock()

		if !ok {
			f, err := fs.doOpenFile(ctx, name, flag, perm)
			if err != nil {
				return nil, err
			}
			dc, ok := f.(*dir)
			if ok {
				fs.m.Lock()
				fs.dirCache[key] = dc
				fs.m.Unlock()
			}
			return dc, nil
		} else {
			dc.lastUsed = time.Now()
			return dc, nil
		}
	} else {
		fs.m.Lock()
		fc, ok := fs.contentCache[key]
		fs.m.Unlock()
		if !ok {
			f, err := fs.doOpenFile(ctx, name, flag, perm)
			if err != nil {
				return nil, err
			}
			fr, ok := f.(*fileRead)
			if ok {
				fs.m.Lock()
				fs.contentCache[key] = fr.content
				fs.m.Unlock()
			}
			return f, nil
		} else {
			fc.lastUsed = time.Now()
			f := &fileRead{
				content: fc,
			}
			return f, nil
		}
	}
}

func (fs *FileSystem) doOpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	isCreate := flag == os.O_RDWR|os.O_CREATE|os.O_TRUNC
	isRead := flag == os.O_RDONLY

	if isRead {
		st, err := fs.Stat(ctx, name)
		if err != nil {
			return nil, err
		}

		if st.IsDir() {
			d := &dir{
				fs:       fs,
				prefix:   name,
				lastUsed: time.Now(),
			}
			return d, nil
		}

		st2, ok := st.(*fileInfo)
		if !ok {
			return nil, fmt.Errorf("not a fileStat")
		}

		fc := newFileContent(fs, st2.oi)
		return &fileRead{
			content: fc,
		}, nil
	} else if isCreate {
		f := &fileWrite{
			fs:  fs,
			key: name,
		}
		err := f.Start()
		if err != nil {
			return nil, err
		}
		return f, nil
	} else {
		return nil, fmt.Errorf("unsupporrted OpenFile invocation: 0x%x", flag)
	}
}

func (fs *FileSystem) delete(ctx context.Context, key string) error {
	fs.forgetCache(key, true)

	slog.Debug("delete", slog.Any("key", key))
	_, err := fs.client2.S3ProxyDeleteObject(ctx, fs.volumeProviderId, models.S3ProxyDeleteObjectRequest{
		Key: key,
	})
	if err != nil {
		return err
	}
	return nil
}

func (fs *FileSystem) listRecursive(ctx context.Context, name string, ret *[]string) error {
	f, err := fs.OpenFile(ctx, name, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return err
	}

	if !st.IsDir() {
		*ret = append(*ret, name)
		return nil
	} else {
		dis, err := f.Readdir(0)
		if err != nil {
			return err
		}
		for _, di := range dis {
			err = fs.listRecursive(ctx, path.Join(name, di.Name()), ret)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (fs *FileSystem) RemoveAll(ctx context.Context, name string) error {
	var files []string
	err := fs.listRecursive(ctx, name, &files)
	if err != nil {
		return err
	}

	for _, f := range files {
		err = fs.delete(ctx, f)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fs *FileSystem) Rename(ctx context.Context, oldName, newName string) error {
	f, err := fs.OpenFile(ctx, oldName, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return err
	}

	if st.IsDir() {
		return fmt.Errorf("renaming dirs not supported")
	}

	trimmedOldName := strings.TrimPrefix(oldName, "/")
	trimmedNewName := strings.TrimPrefix(newName, "/")

	slog.Debug("rename", slog.Any("oldName", oldName), slog.Any("newName", newName))
	_, err = fs.client2.S3ProxyRenameObject(ctx, fs.volumeProviderId, models.S3ProxyRenameObjectRequest{
		OldKey: trimmedOldName,
		NewKey: trimmedNewName,
	})
	fs.forgetCache(trimmedOldName, true)
	fs.forgetCache(trimmedNewName, true)
	if err != nil {
		return err
	}
	return nil
}

func (fs *FileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	key := normalizeName(name)

	if key == "" {
		return &dirInfo{}, nil
	}

	parent := path.Dir(key)
	base := path.Base(key)

	if parent == "." {
		parent = ""
	}

	fs.m.Lock()
	d, ok := fs.dirCache[parent]
	fs.m.Unlock()

	doRet := func(d *dir) (os.FileInfo, error) {
		d.m.Lock()
		defer d.m.Unlock()
		for _, fi := range d.fis {
			if fi.Name() == base {
				return fi, nil
			}
		}
		return nil, os.ErrNotExist
	}

	if ok {
		return doRet(d)
	}

	d = &dir{
		fs:     fs,
		prefix: parent,
	}

	fs.m.Lock()
	fs.dirCache[parent] = d
	fs.m.Unlock()

	_, err := d.Readdir(0)
	if err != nil {
		return nil, err
	}

	return doRet(d)
}

func (fs *FileSystem) forgetCache(key string, lock bool) {
	if lock {
		fs.m.Lock()
		defer fs.m.Unlock()
	}

	key = normalizeName(key)

	delete(fs.contentCache, key)
	delete(fs.dirCache, key)

	if key == "" {
		return
	}

	parentKey := path.Dir(key)
	if parentKey == "." {
		parentKey = ""
	}

	fs.forgetCache(parentKey, false)
}

func (fs *FileSystem) cleanupCache() {
	now := time.Now()

	for k, d := range fs.dirCache {
		anyExpired := false
		for _, st := range d.fis {
			st2, ok := st.(*fileInfo)
			if ok && now.After(st2.oi.PresignedGetUrlExpires.Add(-time.Second*15)) {
				anyExpired = true
				break
			}
		}
		if anyExpired {
			delete(fs.dirCache, k)
			slog.Debug("removing dir cache", slog.Any("key", k))
		}
	}

	for k, cf := range fs.contentCache {
		if now.Before(cf.lastUsed.Add(60 * time.Second)) {
			continue
		}
		slog.Debug("removing content cache", slog.Any("key", k))

		delete(fs.contentCache, k)
	}
}

func normalizeName(n string) string {
	n = strings.TrimPrefix(n, "/")
	n = strings.TrimSuffix(n, "/")
	return n
}
