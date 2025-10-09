//go:build linux

package run_sandbox

import (
	"context"
	"log/slog"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/runner/logs"
)

func (rn *RunSandbox) initFileLogging(ctx context.Context, sandboxDir string, logHandler *logs.MultiLogHandler) error {
	logFile := filepath.Join(sandboxDir, "logs", "run-box.log")
	logWriter := logs.BuildRotatingLogger(logFile)

	logHandler.AddWriter(logWriter)

	return nil
}

func (rn *RunSandbox) initLogsPublishing(ctx context.Context, sandboxDir string) error {
	if rn.Client == nil {
		slog.InfoContext(ctx, "skipping logs publishing (only supported with dboxed api)")
		return nil
	}

	logsDir := filepath.Join(sandboxDir, "logs")

	err := rn.logsPublisher.Start(ctx, rn.Client, rn.BoxId, filepath.Join(logsDir, consts.LogsTailDbFilename))
	if err != nil {
		return err
	}

	err = rn.logsPublisher.PublishDboxedLogsDir(logsDir)
	if err != nil {
		return err
	}

	return nil
}

func (rn *RunSandbox) initLogsPublishingSandbox(ctx context.Context, sandboxDir string) error {
	if rn.Client == nil {
		return nil
	}

	logsDir := filepath.Join(sandboxDir, "logs")
	containersDir := filepath.Join(sandboxDir, "containers")

	err := rn.logsPublisher.PublishMultilogLogsDir(filepath.Join(logsDir, "s6"))
	if err != nil {
		return err
	}

	err = rn.logsPublisher.PublishDockerContainerLogsDir(containersDir)
	if err != nil {
		return err
	}
	return nil
}
