package commands

import (
	"context"
	"github.com/koobox/unboxed/cmd/unboxed/flags"
	"github.com/koobox/unboxed/pkg/sandbox"
	"os"
	"path/filepath"
)

type RuncCmd struct {
	BoxName string `help:"Specify the box name" required:"" arg:""`

	Args []string `arg:"" optional:"" passthrough:""`
}

func (cmd *RuncCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	sandboxDir := filepath.Join(g.WorkDir, "boxes", cmd.BoxName)

	c, err := sandbox.BuildRuncCmd(ctx, sandboxDir, cmd.Args...)
	if err != nil {
		return err
	}
	c.Stdin = os.Stdin

	return c.Run()
}
