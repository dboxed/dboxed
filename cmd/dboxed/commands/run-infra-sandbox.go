package commands

import (
	"context"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	run_infra_sandbox "github.com/dboxed/dboxed/pkg/run-infra-sandbox"
)

type RunInfraSandboxCmd struct {
}

func (cmd *RunInfraSandboxCmd) Run(g *flags.GlobalFlags) error {
	runInfra := run_infra_sandbox.RunInfraSandbox{}

	ctx := context.Background()
	runInfra.Run(ctx)
	// should not exit
	return nil
}
