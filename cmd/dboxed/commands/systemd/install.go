//go:build linux

package systemd

import (
	"context"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/box"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/systemd"
	"github.com/dboxed/dboxed/pkg/util"
)

type SystemdInstallCmd struct {
	Box       string  `help:"Specify box name or id" required:"" arg:""`
	LocalName *string `help:"Override local box name. Defaults to the box name"`
}

func (cmd *SystemdInstallCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient()
	if err != nil {
		return err
	}

	b, err := box.GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	localName := b.Name
	if cmd.LocalName != nil {
		localName = *cmd.LocalName
	}
	err = util.CheckName(localName)
	if err != nil {
		return err
	}

	s := systemd.SystemdInstall{
		ClientAuth: c.GetClientAuth(),
		Box:        b,
		LocalName:  localName,
	}

	return s.Run(ctx)
}
