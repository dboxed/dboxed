package commands

import (
	"context"
	"github.com/koobox/unboxed/cmd/unboxed/flags"
	run_infra_sandbox "github.com/koobox/unboxed/pkg/run-infra-sandbox"
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
