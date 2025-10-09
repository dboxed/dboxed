//go:build linux

package sandbox

import (
	"context"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/run-in-sandbox"
)

type RunInSandbox struct {
}

func (cmd *RunInSandbox) Run(g *flags.GlobalFlags) error {
	runBox := run_in_sandbox.RunInSandbox{}

	ctx := context.Background()
	err := runBox.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}
