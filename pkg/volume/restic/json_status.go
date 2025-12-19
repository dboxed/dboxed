package restic

import (
	"context"
	"encoding/json"
	"log/slog"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

type resticJsonStatusHandler struct {
	curStatus  atomic.Pointer[map[string]any]
	curSummary atomic.Pointer[map[string]any]

	handler func(line string) bool
	stop    func()
}

func buildResticJsonStatusHandler(ctx context.Context) *resticJsonStatusHandler {
	h := &resticJsonStatusHandler{}

	var printMarshalWarningOnce sync.Once
	var printUnknownTypeOnce sync.Once
	h.handler = func(line string) bool {
		var m map[string]any
		err := json.Unmarshal([]byte(line), &m)
		if err != nil {
			printMarshalWarningOnce.Do(func() {
				slog.WarnContext(ctx, "failed to unmarshal restic status line", "error", err)
			})
			return true
		}
		message_type, ok := m["message_type"]
		if !ok {
			printUnknownTypeOnce.Do(func() {
				slog.WarnContext(ctx, "missing message_type field", "line", line)
			})
			return true
		}
		switch message_type {
		case "status":
			h.curStatus.Store(&m)
		case "summary":
			h.curSummary.Store(&m)
		default:
			printUnknownTypeOnce.Do(func() {
				slog.WarnContext(ctx, "unknown message_type", "line", line)
			})
		}
		return true
	}

	var printedStatus map[string]any
	printStatus := func(final bool) {
		status := h.curStatus.Load()
		if status != nil {
			if !reflect.DeepEqual(printedStatus, *status) {
				slog.InfoContext(ctx, "restic status", "status", status)
				printedStatus = *status
			}
		}
		if final {
			if status == nil {
				slog.InfoContext(ctx, "restic did not provide status")
			}
			summary := h.curSummary.Load()
			if summary != nil {
				slog.InfoContext(ctx, "restic summary", "summary", summary)
			}
		}
	}

	done := make(chan struct{})
	go func() {
		for {
			printStatus(false)

			select {
			case <-time.After(1 * time.Second):
			case <-done:
				printStatus(true)
				close(done)
				return
			case <-ctx.Done():
				printStatus(true)
				return
			}
		}
	}()

	h.stop = func() {
		done <- struct{}{}
		<-done
	}

	return h
}
