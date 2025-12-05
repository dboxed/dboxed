package logs

import (
	"io"
	"log/slog"
	"slices"
	"sync/atomic"
)

type MultiLogHandler struct {
	slog.Handler

	writers []io.Writer
	writer  *replaceableWriter
}

type replaceableWriter struct {
	w atomic.Value
}

func (w *replaceableWriter) Write(p []byte) (n int, err error) {
	w2 := w.w.Load().(io.Writer)
	return w2.Write(p)
}

func NewMultiLogHandler(logLevel slog.Level) *MultiLogHandler {
	w := &replaceableWriter{}
	w.w.Store(io.MultiWriter())

	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: logLevel,
	})
	return &MultiLogHandler{
		Handler: h,
		writer:  w,
	}
}

func (h *MultiLogHandler) AddWriter(w io.Writer) {
	h.writers = append(h.writers, w)
	newWriter := io.MultiWriter(h.writers...)
	h.writer.w.Store(newWriter)
}

func (h *MultiLogHandler) RemoveWriter(w io.Writer) {
	h.writers = slices.DeleteFunc(h.writers, func(writer io.Writer) bool {
		return writer == w
	})
	newWriter := io.MultiWriter(h.writers...)
	h.writer.w.Store(newWriter)
}
