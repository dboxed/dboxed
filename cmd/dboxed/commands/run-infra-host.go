package commands

import (
	"context"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	run_infra_host "github.com/dboxed/dboxed/pkg/run-infra-host"
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
