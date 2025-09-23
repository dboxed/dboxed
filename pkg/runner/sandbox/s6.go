//go:build linux

package sandbox

import (
	"context"
	"path/filepath"
)

func (rn *Sandbox) s6SvcCmd(ctx context.Context, serviceName string, args ...string) error {
	args2 := []string{
		"exec", "sandbox", "/command/s6-svc",
	}
	args2 = append(args2, args...)
	args2 = append(args2, filepath.Join("/run/service", serviceName))
	_, err := RunRunc(ctx, rn.SandboxDir, false, args2...)
	if err != nil {
		return err
	}
	return nil
}

func (rn *Sandbox) S6SvcUp(ctx context.Context, serviceName string) error {
	return rn.s6SvcCmd(ctx, serviceName, "-u")
}

func (rn *Sandbox) S6SvcDown(ctx context.Context, serviceName string) error {
	return rn.s6SvcCmd(ctx, serviceName, "-d")
}

func (rn *Sandbox) S6SvcRestart(ctx context.Context, serviceName string) error {
	return rn.s6SvcCmd(ctx, serviceName, "-r")
}
