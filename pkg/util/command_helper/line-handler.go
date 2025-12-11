package command_helper

import (
	"bufio"
	"io"
	"time"
)

type LineHandler struct {
	handler func(line string)

	r io.ReadCloser
	w io.WriteCloser
	s *bufio.Scanner

	done chan struct{}
}

type Line struct {
	Line string    `json:"line"`
	Time time.Time `json:"time"`
}

func NewLineHandler(handler func(line string)) *LineHandler {
	r, w := io.Pipe()
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, bufio.MaxScanTokenSize), 1024*1024)

	l := &LineHandler{
		handler: handler,
		r:       r,
		w:       w,
		s:       s,
		done:    make(chan struct{}),
	}
	l.start()
	return l
}

func (lh *LineHandler) Close() error {
	err := lh.w.Close()
	if err != nil {
		return err
	}
	return lh.Wait()
}

func (lh *LineHandler) Wait() error {
	<-lh.done
	return lh.s.Err()
}

func (lh *LineHandler) start() {
	go func() {
		defer close(lh.done)
		for lh.s.Scan() {
			lh.handler(lh.s.Text())
		}
	}()
}

func (lh *LineHandler) Write(b []byte) (n int, err error) {
	return lh.w.Write(b)
}
