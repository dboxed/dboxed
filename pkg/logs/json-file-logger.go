package logs

import (
	"bufio"
	"bytes"
	"github.com/koobox/unboxed/pkg/logs/jsonlog"
	"io"
	"sync"
	"time"
)

type JsonFileLogger struct {
	Out    io.Writer
	Stream string

	r     io.Reader
	w     io.Writer
	s     *bufio.Scanner
	errCh chan error

	pendingLines chan *bytes.Buffer

	p  *sync.Pool
	wg sync.WaitGroup
}

func NewJsonFileLogger(out io.Writer, stream string) *JsonFileLogger {
	r, w := io.Pipe()
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, bufio.MaxScanTokenSize), 1024*1024)

	p := &sync.Pool{
		New: func() any {
			return bytes.NewBuffer(make([]byte, 0, 256))
		},
	}

	l := &JsonFileLogger{
		Out:          out,
		Stream:       stream,
		r:            r,
		w:            w,
		s:            s,
		errCh:        make(chan error),
		pendingLines: make(chan *bytes.Buffer),
		p:            p,
	}
	l.start()
	return l
}

func (l *JsonFileLogger) Wait() error {
	l.wg.Wait()
	err := <-l.errCh
	return err
}

func (l *JsonFileLogger) start() {
	l.wg.Add(2)
	go func() {
		defer l.wg.Done()
		for l.s.Scan() {
			l.queueJsonLine(l.s.Bytes())
		}
		l.errCh <- l.s.Err()
		close(l.pendingLines)
	}()
	go func() {
		defer l.wg.Done()
		for e := range l.pendingLines {
			l.writeJsonLine(e)
		}
	}()
}

func (l *JsonFileLogger) Write(b []byte) (n int, err error) {
	return l.w.Write(b)
}

func (l *JsonFileLogger) queueJsonLine(line []byte) {
	b := l.p.Get().(*bytes.Buffer)
	b.Reset()

	j := jsonlog.JSONLogs{
		Log:     line,
		Stream:  l.Stream,
		Created: time.Now(),
	}
	err := j.MarshalJSONBuf(b)
	if err != nil {
		panic(err)
	}
	b.WriteByte('\n')

	l.pendingLines <- b
}

func (l *JsonFileLogger) writeJsonLine(b *bytes.Buffer) {
	_, _ = l.Out.Write(b.Bytes())
	l.p.Put(b)
}
