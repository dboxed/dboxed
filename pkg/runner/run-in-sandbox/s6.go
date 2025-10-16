package run_in_sandbox

import (
	"context"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/util"
)

func (rn *RunInSandbox) s6SvcCmd(ctx context.Context, serviceName string, args ...string) error {
	args2 := []string{}
	args2 = append(args2, args...)
	args2 = append(args2, filepath.Join("/run/service", serviceName))

	c := util.CommandHelper{
		Command: "/command/s6-svc",
		Args:    args,
		LogCmd:  true,
	}

	err := c.Run(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (rn *RunInSandbox) S6SvcUp(ctx context.Context, serviceName string) error {
	return rn.s6SvcCmd(ctx, serviceName, "-u", "-wu")
}

func (rn *RunInSandbox) S6SvcDown(ctx context.Context, serviceName string) error {
	return rn.s6SvcCmd(ctx, serviceName, "-d", "-wd")
}

func (rn *RunInSandbox) S6SvcRestart(ctx context.Context, serviceName string) error {
	return rn.s6SvcCmd(ctx, serviceName, "-r", "-wr")
}
