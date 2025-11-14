package run_in_sandbox

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/runner/logs"
)

func (rn *RunInSandbox) initLogsPublishing(ctx context.Context) error {
	if rn.client == nil {
		slog.InfoContext(ctx, "skipping logs publishing (only supported with dboxed api)")
		return nil
	}

	tta, err := logs.NewTailToApi(ctx, rn.client, filepath.Join(consts.LogsDir, consts.LogsTailDbFilename), rn.sandboxInfo.Box.ID)
	if err != nil {
		return err
	}

	err = rn.logsPublisher.Start(ctx, tta.MultiTail)
	if err != nil {
		return err
	}

	err = rn.logsPublisher.PublishDboxedLogsDir(consts.LogsDir)
	if err != nil {
		return err
	}

	err = rn.logsPublisher.PublishS6Logs(filepath.Join(consts.LogsDir, "s6"))
	if err != nil {
		return err
	}

	err = rn.logsPublisher.PublishVolumeServiceLogs(consts.VolumesDir)
	if err != nil {
		return err
	}

	err = rn.logsPublisher.PublishDockerContainerLogsDir(consts.ContainersDir)
	if err != nil {
		return err
	}
	return nil
}

func (rn *RunInSandbox) buildReconcileLogger() (*slog.Logger, io.WriteCloser) {
	logFile := logs.BuildRotatingLogger(filepath.Join(consts.LogsDir, "reconcile.log"))
	defer logFile.Close()
	h := slog.NewJSONHandler(logFile, &slog.HandlerOptions{})
	log := slog.New(h)
	return log, logFile
}
