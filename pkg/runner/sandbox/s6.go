//go:build linux

package sandbox

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/opencontainers/runc/libcontainer"
)

func (rn *Sandbox) s6SvcCmd(ctx context.Context, serviceName string, args ...string) error {
	c, err := rn.GetSandboxContainer()
	if err != nil {
		return err
	}

	args2 := []string{
		"/command/s6-svc",
	}
	args2 = append(args2, args...)
	args2 = append(args2, filepath.Join("/run/service", serviceName))

	p := &libcontainer.Process{
		Args:   args2,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	err = c.Run(p)
	if err != nil {
		return err
	}
	ps, err := p.Wait()
	if err != nil {
		return err
	}
	if ps.ExitCode() != 0 {
		return fmt.Errorf("s6-svc exited with exit code %d", ps.ExitCode())
	}
	return nil
}

func (rn *Sandbox) S6SvcUp(ctx context.Context, serviceName string) error {
	return rn.s6SvcCmd(ctx, serviceName, "-u", "-wu")
}

func (rn *Sandbox) S6SvcDown(ctx context.Context, serviceName string) error {
	return rn.s6SvcCmd(ctx, serviceName, "-d", "-wd")
}

func (rn *Sandbox) S6SvcRestart(ctx context.Context, serviceName string) error {
	return rn.s6SvcCmd(ctx, serviceName, "-r", "-wr")
}
