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
	ctx := context.Background()

	// this will respect the DBOXED_SANDBOX=1 variable and load the auth config from consts.BoxClientAuthFile
	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	runBox := run_in_sandbox.RunInSandbox{
		Client: c,
	}

	err = runBox.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}
