package multitail

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/fsnotify/fsnotify"
	"github.com/gofrs/flock"
)

type MultiTail struct {
	ctx context.Context
	fl  *flock.Flock
	bdb *badger.DB

	opts MultiTailOptions

	tails     map[string]*Tail
	watchers  map[string]*fsnotify.Watcher
	doneGroup sync.WaitGroup
	stopped   bool
	m         sync.Mutex
}

type MultiTailOptions struct {
	LineBatchSize    int
	LineBatchLinger  time.Duration
	LineBatchHandler LineBatchHandler
}

type LogMetadata struct {
	FileName string         `json:"fileName"`
	Format   string         `json:"format"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type LineBatchHandler func(metadata LogMetadata, lines []*Line) error
type BuildMetadataFunc func(path string) (LogMetadata, error)

type dbFileEntry struct {
	Inode  uint64 `json:"inode"`
	Offset int64  `json:"offset"`
}

func NewMultiTail(ctx context.Context, tailDbFile string, opts MultiTailOptions) (*MultiTail, error) {
	fl := flock.New(tailDbFile + ".multitail-lock")

	slog.InfoContext(ctx, "multitail: waiting for db lock")
	err := fl.Lock()
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "multitail: db lock acquired")

	bopts := badger.
		DefaultOptions(tailDbFile).
		WithLoggingLevel(badger.WARNING)

	bdb, err := badger.Open(bopts)
	if err != nil {
		return nil, err
	}
	return &MultiTail{
		ctx:      ctx,
		fl:       fl,
		bdb:      bdb,
		opts:     opts,
		tails:    map[string]*Tail{},
		watchers: map[string]*fsnotify.Watcher{},
	}, nil
}

func (mt *MultiTail) StopAndWait() {
	slog.Info("multitail: stopping...")

	mt.m.Lock()
	mt.stopped = true
	for _, w := range mt.watchers {
		_ = w.Close()
	}
	for _, t := range mt.tails {
		t.Stop()
	}
	mt.m.Unlock()

	slog.Info("multitail: waiting for routines to finish")
	mt.doneGroup.Wait()

	slog.Info("multitail: closing db")
	_ = mt.bdb.Close()

	slog.Info("multitail: releasing db lock")
	_ = mt.fl.Unlock()
}

func (mt *MultiTail) TailFile(path string, metadata LogMetadata) error {
	mt.m.Lock()
	defer mt.m.Unlock()

	if mt.stopped {
		return fmt.Errorf("multitail already stopped")
	}

	if _, ok := mt.tails[path]; ok {
		return nil
	}

	slog.Info("multitail: tailing new path", slog.Any("path", path), slog.Any("fileName", metadata.FileName))

	var inode uint64
	var offset int64

	err := mt.bdb.View(func(txn *badger.Txn) error {
		e, err := mt.getDbFileEntry(txn, metadata.FileName)
		if err != nil {
			return err
		}
		if e == nil {
			return nil
		}
		inode = e.Inode
		offset = e.Offset
		slog.Info("multitail: using stored offset", slog.Any("inode", inode),
			slog.Any("offset", offset), slog.Any("fileName", metadata.FileName))
		return nil
	})
	if err != nil {
		return err
	}

	tf, err := NewTail(mt.ctx, path, TailOptions{
		Inode:  inode,
		Offset: offset,
	})
	if err != nil {
		return err
	}

	mt.tails[path] = tf

	mt.doneGroup.Add(1)
	go func() {
		defer func() {
			mt.doneGroup.Done()
		}()
		mt.runHandleTail(tf, path, metadata)
	}()

	return nil
}

func (mt *MultiTail) WatchDir(dir string, pattern string, watchDepth int,
	buildMetadata BuildMetadataFunc) error {
	mt.m.Lock()
	defer mt.m.Unlock()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	mt.watchers[dir] = watcher

	mt.doneGroup.Add(1)
	go func() {
		defer mt.doneGroup.Done()

		err = mt.handleWatchedPath(watcher, dir, dir, nil, pattern, watchDepth, buildMetadata)
		if err != nil {
			slog.Error("error in handleWatchedPath", slog.Any("error", err))
			return
		}

		for e := range watcher.Events {
			if e.Has(fsnotify.Remove) || e.Has(fsnotify.Rename) {
				mt.m.Lock()
				tf := mt.tails[e.Name]
				if tf != nil {
					slog.Info("multitail: stopping tail for file", slog.Any("path", e.Name))

					tf.Stop()
					delete(mt.tails, e.Name)
				}
				mt.m.Unlock()
				_ = mt.bdb.Update(func(txn *badger.Txn) error {
					return txn.Delete([]byte(e.Name))
				})
			}
			if e.Has(fsnotify.Create) {
				err := mt.handleWatchedPath(watcher, dir, e.Name, nil, pattern, watchDepth, buildMetadata)
				if err != nil {
					if !os.IsNotExist(err) {
						slog.Error("error in handleWatchedPath", slog.Any("error", err))
					}
					continue
				}
			}
		}
	}()

	return nil
}

func (mt *MultiTail) handleWatchedPath(
	watcher *fsnotify.Watcher, rootDir string, path string, st os.FileInfo,
	pattern string, watchDepth int, buildMetadata BuildMetadataFunc) error {
	slog.Debug("multitail: handleWatchedPath", slog.Any("path", path), slog.Any("rootDir", rootDir))

	if st == nil {
		var err error
		st, err = os.Stat(path)
		if err != nil {
			return err
		}
	}

	relPath, err := filepath.Rel(rootDir, path)
	if err != nil {
		return err
	}

	if st.IsDir() {
		depth := len(strings.Split(relPath, string(filepath.Separator)))
		if relPath != "." && depth > watchDepth {
			return nil
		}
		slog.Debug("multitail: starting to watch new dir", slog.Any("path", path))
		err = watcher.Add(path)
		if err != nil {
			return err
		}
		fes, err := os.ReadDir(path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		for _, fe := range fes {
			st, err := fe.Info()
			if err != nil {
				return err
			}
			err = mt.handleWatchedPath(watcher, rootDir, filepath.Join(path, fe.Name()), st, pattern, watchDepth, buildMetadata)
			if err != nil {
				return err
			}
		}
	} else {
		slog.Debug("multitail: checking pattern", slog.Any("relPath", relPath), slog.Any("pattern", pattern))
		m, err := filepath.Match(pattern, relPath)
		if err != nil {
			return err
		}
		if m {
			metadata, err := buildMetadata(path)
			if err != nil {
				return err
			}
			err = mt.TailFile(path, metadata)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (mt *MultiTail) runHandleTail(tf *Tail, path string, metadata LogMetadata) {
	var curBatch []*Line

	tryHandleBatch := func() {
		if len(curBatch) != 0 {
			lastLine := curBatch[len(curBatch)-1]
			mt.handleLineBatch(metadata, curBatch)
			curBatch = curBatch[0:0]

			err := mt.bdb.Update(func(txn *badger.Txn) error {
				return mt.setDbFileEntry(txn, metadata.FileName, dbFileEntry{
					Offset: lastLine.Offset,
					Inode:  tf.Inode,
				})
			})
			if err != nil {
				slog.Error("failed to update tail db", slog.Any("path", path), slog.Any("error", err))
			}
		}
	}

loop:
	for {
		select {
		case newLine, ok := <-tf.Lines:
			if !ok {
				tryHandleBatch()
				break loop
			}
			curBatch = append(curBatch, newLine)
			if len(curBatch) >= mt.opts.LineBatchSize {
				tryHandleBatch()
			}
		case <-time.After(mt.opts.LineBatchLinger):
			tryHandleBatch()
		}
	}
}

func (mt *MultiTail) handleLineBatch(metadata LogMetadata, lines []*Line) {
	for {
		err := mt.opts.LineBatchHandler(metadata, lines)
		if err == nil {
			return
		}
		slog.Info("handleLineBatch failed, retrying", slog.Any("error", err))
		time.Sleep(1 * time.Second)
	}
}

func (mt *MultiTail) getDbFileEntry(txn *badger.Txn, file string) (*dbFileEntry, error) {
	item, err := txn.Get([]byte(file))
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, nil
		}
		return nil, err
	}
	b, err := item.ValueCopy(nil)
	if err != nil {
		return nil, err
	}
	var ret dbFileEntry
	err = json.Unmarshal(b, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (mt *MultiTail) setDbFileEntry(txn *badger.Txn, file string, e dbFileEntry) error {
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return txn.Set([]byte(file), b)
}

func getInode(st os.FileInfo) (uint64, error) {
	st2, ok := st.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, fmt.Errorf("not a syscall.Stat_t")
	}
	return st2.Ino, nil
}
