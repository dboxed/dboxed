package commands

import (
	"context"
	"github.com/koobox/unboxed/cmd/unboxed/flags"
	run_infra_host "github.com/koobox/unboxed/pkg/run-infra-host"
)

type RunInfraHostCmd struct {
}

func (cmd *RunInfraHostCmd) Run(g *flags.GlobalFlags) error {
	runInfra := run_infra_host.RunInfraHost{}

	ctx := context.Background()
	runInfra.Run(ctx)
	// should not exit
	return nil
}
