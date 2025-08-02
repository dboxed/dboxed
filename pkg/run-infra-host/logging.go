package run_infra_host

import (
	"github.com/dboxed/dboxed/pkg/logs"
	"github.com/dboxed/dboxed/pkg/types"
	"log/slog"
	"path/filepath"
)

func (rn *RunInfraHost) initLogging() {
	infraLog := logs.BuildRotatingLogger(filepath.Join(types.LogsDir, "infra-host.log"))

	handler := slog.NewJSONHandler(infraLog, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))
}

func (rn *RunInfraHost) stopLogging() {
}
