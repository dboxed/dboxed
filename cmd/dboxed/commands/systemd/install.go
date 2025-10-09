//go:build linux

package systemd

import (
	"context"
	"fmt"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/systemd"
	"github.com/dboxed/dboxed/pkg/util"
)

type SystemdInstallCmd struct {
	Box       string  `help:"Specify box name or id" required:"" arg:""`
	LocalName *string `help:"Override local box name. Defaults to the box <name>-<uuid>"`
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

	localName := fmt.Sprintf("%s-%s", box.Name, box.Uuid)
	if cmd.LocalName != nil {
		localName = *cmd.LocalName
	}
	err = util.CheckName(localName)
	if err != nil {
		return err
	}

	s := systemd.SystemdInstall{
		ClientAuth: c.GetClientAuth(true),
		Box:        box,
		LocalName:  localName,
	}

	return s.Run(ctx)
}
