//go:build linux

package sandbox

import (
	"context"
	"path/filepath"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/sandbox"
)

type KillCmd struct {
	SandboxName string `help:"Specify the local sandbox name" required:"" arg:""`

	Signal *string `help:"Specify the signal to be sent to the init process"`
}

func (cmd *KillCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	sandboxDir := filepath.Join(g.WorkDir, "boxes", cmd.SandboxName)

	args := []string{
		"kill",
		"sandbox",
	}
	if cmd.Signal != nil {
		args = append(args, *cmd.Signal)
	}

	c := sandbox.BuildRuncCmd(ctx, sandboxDir, args...)

	return c.Run()
}
