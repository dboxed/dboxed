package commands

import (
	"context"
	"github.com/koobox/unboxed/cmd/unboxed/flags"
	run_infra "github.com/koobox/unboxed/pkg/run-infra-sandbox"
	"time"
)

type RunInfraSandboxCmd struct {
}

func (cmd *RunInfraSandboxCmd) Run(g *flags.GlobalFlags) error {
	runInfraSandbox := run_infra.RunInfraSandbox{}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := runInfraSandbox.Start(ctx)
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
