//go:build !linux

package command_helper

import (
	"context"
	"fmt"
	"io"
)

type ContainerHolder struct {
}

func (c *CommandHelper) isContainer() bool {
	return false
}

func (c *CommandHelper) runContainer(ctx context.Context, stdout io.Writer, stderr io.Writer) error {
	return fmt.Errorf("runContainer only supported on linux")
}
