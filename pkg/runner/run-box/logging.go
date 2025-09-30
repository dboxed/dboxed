//go:build linux

package run_box

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/runner/logs"
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

func (rn *RunBox) initLogsPublishing(ctx context.Context, sandboxDir string, boxSpec *boxspec.BoxSpec) error {
	//if rn.natsConn == nil {
	//	slog.InfoContext(ctx, "skipping logs publishing (only supported with nats)")
	//	return nil
	//}
	//
	//logsDir := filepath.Join(sandboxDir, "logs")
	//
	//err := rn.logsPublisher.Start(ctx, rn.natsConn, boxSpec, filepath.Join(logsDir, consts.LogsTailDbFilename))
	//if err != nil {
	//	return err
	//}
	//
	//err = rn.logsPublisher.PublishDboxedLogsDir(logsDir)
	//if err != nil {
	//	return err
	//}

	return nil
}

func (rn *RunBox) initLogsPublishingSandbox(ctx context.Context, sandboxDir string, boxSpec *boxspec.BoxSpec) error {
	//if rn.natsConn == nil {
	//	return nil
	//}
	//
	//logsDir := filepath.Join(sandboxDir, "logs")
	//containersDir := filepath.Join(sandboxDir, "containers")
	//
	//err := rn.logsPublisher.PublishMultilogLogsDir(filepath.Join(logsDir, "s6"))
	//if err != nil {
	//	return err
	//}
	//
	//err = rn.logsPublisher.PublishDockerContainerLogsDir(containersDir)
	//if err != nil {
	//	return err
	//}
	return nil
}
