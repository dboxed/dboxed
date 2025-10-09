package logs

import (
	"io"
	"log/slog"
	"sync/atomic"
)

type MultiLogHandler struct {
	slog.Handler

	writer *replaceableWriter
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
	oldWriter := h.writer.w.Load().(io.Writer)
	newWriter := io.MultiWriter(oldWriter, w)
	h.writer.w.Store(newWriter)
}
