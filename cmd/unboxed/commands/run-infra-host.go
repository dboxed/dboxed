package commands

import (
	"context"
	"github.com/koobox/unboxed/cmd/unboxed/flags"
	run_infra "github.com/koobox/unboxed/pkg/run-infra-host"
	"time"
)

type RunInfraHostCmd struct {
}

func (cmd *RunInfraHostCmd) Run(g *flags.GlobalFlags) error {
	runInfraHost := run_infra.RunInfraHost{}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := runInfraHost.Start(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
