package runner

import (
	"context"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/systemd"
)

type SystemdCmd struct {
	Install SystemdInstallCmd `cmd:"" help:"Install dboxed as a systemd service"`
}

type SystemdInstallCmd struct {
	flags.BoxSourceFlags

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
		Nkey:    cmd.Nkey,
		BoxName: cmd.BoxName,
	}

	return s.Run(ctx)
}
