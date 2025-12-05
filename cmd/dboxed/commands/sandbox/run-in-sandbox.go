//go:build linux

package sandbox

import (
	"context"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/logs"
	"github.com/dboxed/dboxed/pkg/runner/run-in-sandbox"
)

type RunInSandbox struct {
}

func (cmd *RunInSandbox) Run(g *flags.GlobalFlags, logHandler *logs.MultiLogHandler) error {
	ctx := context.Background()

	runBox := run_in_sandbox.RunInSandbox{
		WorkDir: g.WorkDir,
	}

	err := runBox.Run(ctx, logHandler)
	if err != nil {
		return err
	}

	return nil
}
