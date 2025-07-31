package multitail

import (
	"bufio"
	"context"
	"io"
	"os"
	"time"
)

type Tail struct {
	ctx   context.Context
	Lines chan *Line

	Inode     uint64
	file      *os.File
	stopOnEof chan struct{}
}

type TailOptions struct {
	Inode  uint64
	Offset int64
}

type Line struct {
	Offset int64

	Line string
	Time time.Time
	Err  error
}

func NewTail(ctx context.Context, file string, opts TailOptions) (*Tail, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	doClose := true
	defer func() {
		if doClose {
			_ = f.Close()
		}
	}()

	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	curInode, err := getInode(st)
	if err != nil {
		return nil, err
	}

	if curInode == opts.Inode && opts.Offset != 0 {
		_, err := f.Seek(opts.Offset, io.SeekStart)
		if err != nil {
			return nil, err
		}
	}

	t := &Tail{
		ctx:       ctx,
		Lines:     make(chan *Line),
		file:      f,
		Inode:     curInode,
		stopOnEof: make(chan struct{}),
	}

	go func() {
		defer func() {
			close(t.Lines)
			_ = t.file.Close()
		}()
		err := t.runLineReader()
		if err != nil && err != io.EOF {
			t.Lines <- &Line{
				Time: time.Now(),
				Err:  err,
			}
		}
	}()

	doClose = false
	return t, nil
}

func (t *Tail) Stop() {
	close(t.stopOnEof)
}

func (t *Tail) runLineReader() error {
	pw := &pollWrapper{
		ReadCloser: t.file,
		ctx:        t.ctx,
		stopOnEOF:  t.stopOnEof,
	}
	br := bufio.NewReader(pw)

	getOffset := func() int64 {
		offset, err := t.file.Seek(0, io.SeekCurrent)
		if err != nil {
			return -1
		}
		return offset - int64(br.Buffered())
	}
	handleLine := func(line string, offset int64) {
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		t.Lines <- &Line{
			Offset: offset,
			Line:   line,
			Time:   time.Now(),
		}
	}

	for {
		line, err := br.ReadString('\n')
		offset := getOffset()
		if err != nil {
			if line != "" {
				handleLine(line, offset)
			}
			return err
		}
		handleLine(line, offset)
	}
}
