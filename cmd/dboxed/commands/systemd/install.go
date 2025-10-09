//go:build linux

package systemd

import (
	"context"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/sandbox"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/systemd"
)

type SystemdInstallCmd struct {
	Box         string  `help:"Specify box name or id" required:"" arg:""`
	SandboxName *string `help:"Override local sandbox name. Defaults to the box <name>-<uuid>"`
}

func (cmd *SystemdInstallCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	box, err := commandutils.GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	sandboxName, err := sandbox.GetSandboxName(box, cmd.SandboxName)
	if err != nil {
		return err
	}

	s := systemd.SystemdInstall{
		ClientAuth:  c.GetClientAuth(true),
		Box:         box,
		SandboxName: sandboxName,
	}

	return s.Run(ctx)
}
