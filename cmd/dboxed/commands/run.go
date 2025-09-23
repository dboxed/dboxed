package commands

import (
	"context"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/run-box"
)

type RunCmd struct {
	flags.BoxSourceFlags

	InfraImage string `help:"Specify the infra/sandbox image to use" default:"${default_infra_image}"`

	BoxName     string `help:"Specify the box name" required:""`
	VethCidrArg string `help:"CIDR to use for veth pairs. dboxed will dynamically allocate 2 IPs from this CIDR per box" default:"1.2.3.0/24"`

	WaitBeforeExit *time.Duration `help:"Wait before finally exiting. This gives the process time to print stdout/stderr messages that might be lost. Especially useful in combination with --debug"`
}

func (cmd *RunCmd) Run(g *flags.GlobalFlags) error {
	defer func() {
		if cmd.WaitBeforeExit != nil {
			slog.Info("sleeping before exit")
			time.Sleep(*cmd.WaitBeforeExit)
		}
	}()

	url, err := cmd.GetBoxUrl()
	if err != nil {
		return err
	}

	runBox := run_box.RunBox{
		Debug:           g.Debug,
		InfraImage:      cmd.InfraImage,
		BoxUrl:          url,
		Nkey:            cmd.Nkey,
		BoxName:         cmd.BoxName,
		WorkDir:         g.WorkDir,
		VethNetworkCidr: cmd.VethCidrArg,
	}

	ctx := context.Background()
	err = runBox.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}
