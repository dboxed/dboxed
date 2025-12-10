package multitail

import (
	"context"
	"io"
	"time"
)

type pollWrapper struct {
	io.ReadCloser
	stopOnEOF chan struct{}
	ctx       context.Context
}

func (w *pollWrapper) Read(p []byte) (n int, err error) {
	stopOnEOF := false
	for {
		b, err := w.ReadCloser.Read(p)
		if err == nil {
			return b, nil
		}
		if err != io.EOF {
			return 0, err
		} else if stopOnEOF {
			return 0, io.EOF
		}

		select {
		case <-w.stopOnEOF:
			stopOnEOF = true
		case <-w.ctx.Done():
			return 0, w.ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
}
