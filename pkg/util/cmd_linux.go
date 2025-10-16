//go:build linux

package util

import (
	"context"
	"fmt"
	"io"
	"os"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/opencontainers/runc/libcontainer"
)

type ContainerHolder struct {
	Container   *libcontainer.Container
	ImageConfig *v1.ImageConfig
}

func (c *CommandHelper) isContainer() bool {
	return c.Container != nil
}

func (c *CommandHelper) runContainer(ctx context.Context, stdout io.Writer, stderr io.Writer) error {
	args := []string{c.Command}
	args = append(args, c.Args...)

	var env []string
	if c.ImageConfig != nil {
		env = append(env, c.ImageConfig.Env...)
	}
	env = append(env, c.Env...)

	p := &libcontainer.Process{
		Args:   args,
		Env:    env,
		Stdout: stdout,
		Stderr: stderr,
	}
	err := c.Container.Run(p)
	if err != nil {
		return err
	}

	psCh := make(chan *os.ProcessState, 1)
	errCh := make(chan error, 1)
	go func() {
		ps, err := p.Wait()
		if err != nil {
			errCh <- err
		} else {
			psCh <- ps
		}
	}()

	select {
	case ps := <-psCh:
		if ps.ExitCode() != 0 {
			return fmt.Errorf("command exited with exit code %d", ps.ExitCode())
		}
		return nil
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
