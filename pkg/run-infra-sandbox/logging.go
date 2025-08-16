package run_infra_sandbox

import (
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/dboxed/dboxed/pkg/logs"
	"github.com/dboxed/dboxed/pkg/logs/line_handler"
	"github.com/dboxed/dboxed/pkg/types"
)

func (rn *RunInfraSandbox) initLogging() {
	infraLog := logs.BuildRotatingLogger(filepath.Join(types.LogsDir, "infra-sandbox.log"))
	handler := slog.NewJSONHandler(infraLog, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))

	dockerCliLog := logs.BuildRotatingLogger(filepath.Join(types.LogsDir, "docker-cli.log"))
	dockerCliLogger := slog.New(slog.NewJSONHandler(dockerCliLog, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	dockerCliStdout := line_handler.NewLineHandler(func(line string) {
		line = strings.TrimSuffix(line, "\n")
		dockerCliLogger.Info(line, slog.Any("stream", "stdout"))
	})
	dockerCliStderr := line_handler.NewLineHandler(func(line string) {
		line = strings.TrimSuffix(line, "\n")
		dockerCliLogger.Info(line, slog.Any("stream", "stderr"))
	})

	rn.dockerCliStdout = dockerCliStdout
	rn.dockerCliStderr = dockerCliStderr
}

func (rn *RunInfraSandbox) stopLogging() {
	_ = rn.dockerCliStdout.Close()
	_ = rn.dockerCliStderr.Close()
}
