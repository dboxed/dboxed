package commands

import (
	"context"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/run-box-in-sandbox"
)

type RunBoxInSandbox struct {
}

func (cmd *RunBoxInSandbox) Run(g *flags.GlobalFlags) error {
	runBox := run_box_in_sandbox.RunBoxInSandbox{
		Debug: g.Debug,
	}

	ctx := context.Background()
	err := runBox.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}
