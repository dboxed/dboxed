package start_box

import (
	"context"
	"github.com/dboxed/dboxed/pkg/logs"
	"github.com/dboxed/dboxed/pkg/types"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func (rn *StartBox) initFileLogging(ctx context.Context, sandboxDir string) error {
	logFile := filepath.Join(sandboxDir, "logs", "start-box.log")
	logWriter := logs.BuildRotatingLogger(logFile)
	dupWriter := io.MultiWriter(os.Stderr, logWriter)

	h := slog.NewJSONHandler(dupWriter, &slog.HandlerOptions{})
	log := slog.New(h)
	slog.SetDefault(log)

	return nil
}

func (rn *StartBox) initLogsPublishing(ctx context.Context, sandboxDir string) error {
	logsDir := filepath.Join(sandboxDir, "logs")

	err := rn.logsPublisher.Start(ctx, *rn.boxSpec, filepath.Join(logsDir, types.LogsTailDbFilename))
	if err != nil {
		return err
	}

	err = rn.logsPublisher.PublishDboxedLogsDir(logsDir)
	if err != nil {
		return err
	}

	return nil
}
