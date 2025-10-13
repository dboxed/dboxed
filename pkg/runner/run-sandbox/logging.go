//go:build linux

package run_sandbox

import (
	"context"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/runner/logs"
)

func (rn *RunSandbox) initFileLogging(ctx context.Context, sandboxDir string, logHandler *logs.MultiLogHandler) error {
	logFile := filepath.Join(sandboxDir, "logs", "run-box.log")
	logWriter := logs.BuildRotatingLogger(logFile)

	logHandler.AddWriter(logWriter)

	return nil
}
