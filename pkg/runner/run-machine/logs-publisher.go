//go:build linux

package run_machine

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/runner/logs/multitail"
)

type LogsPublisher struct {
	MachineId string

	mt *multitail.MultiTail
}

func (lp *LogsPublisher) Stop(cancelAfter *time.Duration) {
	if lp == nil {
		return
	}
	if lp.mt != nil {
		lp.mt.StopAndWait(cancelAfter)
	}
}

func (lp *LogsPublisher) Start(ctx context.Context, mt *multitail.MultiTail) error {
	slog.InfoContext(ctx, "initializing logs publishing to dboxed api")
	lp.mt = mt
	return nil
}

func (lp *LogsPublisher) PublishMachineLogsDir(dir string) error {
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	buildMetadata := func(path string) (boxspec.LogMetadata, error) {
		fileName := filepath.Join("dboxed", filepath.Base(path))
		format := "slog-json"
		if strings.HasSuffix(fileName, ".stdout.log") {
			format = "raw"
		}
		return boxspec.LogMetadata{
			MachineId: &lp.MachineId,
			FileName:  fileName,
			Format:    format,
			Metadata:  map[string]any{},
		}, nil
	}

	if lp.mt != nil {
		return lp.mt.WatchDir(dir, "*.log", 0, buildMetadata)
	}
	return nil
}
