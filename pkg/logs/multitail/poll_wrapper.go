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
	for {
		b, err := w.ReadCloser.Read(p)
		if err == nil {
			return b, nil
		}
		if err != io.EOF {
			return 0, err
		}

		select {
		case <-w.stopOnEOF:
			return 0, io.EOF
		case <-w.ctx.Done():
			return 0, w.ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
}
