//go:build linux

package box

import (
	"context"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/sandbox"
)

type RuncCmd struct {
	LocalName string `help:"Specify the local box name" required:"" arg:""`

	Args []string `arg:"" optional:"" passthrough:""`
}

func (cmd *RuncCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	sandboxDir := filepath.Join(g.WorkDir, "boxes", cmd.LocalName)

	c := sandbox.BuildRuncCmd(ctx, sandboxDir, cmd.Args...)
	c.Stdin = os.Stdin

	return c.Run()
}
