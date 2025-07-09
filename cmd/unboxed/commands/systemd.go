package commands

import (
	"context"
	"github.com/koobox/unboxed/cmd/unboxed/flags"
	"github.com/koobox/unboxed/pkg/systemd"
)

type SystemdCmd struct {
	Install SystemdInstallCmd `cmd:"" help:"Install unboxed as a systemd service"`
}

type SystemdInstallCmd struct {
	flags.BoxUrlFlags

	BoxName string `help:"Specify the box name" required:""`
}

func (cmd *SystemdInstallCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	u, err := cmd.GetBoxUrl()
	if err != nil {
		return err
	}

	s := systemd.SystemdInstall{
		BoxUrl:  u,
		BoxName: cmd.BoxName,
	}

	return s.Run(ctx)
}
