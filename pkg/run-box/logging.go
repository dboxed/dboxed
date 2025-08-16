package run_box

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/logs"
	"github.com/dboxed/dboxed/pkg/types"
)

func (rn *RunBox) initFileLogging(ctx context.Context, sandboxDir string) error {
	logFile := filepath.Join(sandboxDir, "logs", "run-box.log")
	logWriter := logs.BuildRotatingLogger(logFile)
	dupWriter := io.MultiWriter(os.Stderr, logWriter)

	h := slog.NewJSONHandler(dupWriter, &slog.HandlerOptions{})
	log := slog.New(h)
	slog.SetDefault(log)

	return nil
}

func (rn *RunBox) initLogsPublishing(ctx context.Context, sandboxDir string) error {
	logsDir := filepath.Join(sandboxDir, "logs")
	dockerDir := filepath.Join(sandboxDir, "docker")

	err := rn.logsPublisher.Start(ctx, *rn.boxSpec, filepath.Join(logsDir, types.LogsTailDbFilename))
	if err != nil {
		return err
	}

	err = rn.logsPublisher.PublishDboxedLogsDir(logsDir)
	if err != nil {
		return err
	}

	err = rn.logsPublisher.PublishDockerLogsDir(dockerDir)
	if err != nil {
		return err
	}
	return nil
}
