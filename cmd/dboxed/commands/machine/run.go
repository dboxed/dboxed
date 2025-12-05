//go:build linux

package machine

import (
	"context"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/logs"
	run_machine "github.com/dboxed/dboxed/pkg/runner/run-machine"
)

type RunCmd struct {
	Machine string `help:"Specify the machine ID or name" required:"" arg:""`

	InfraImage string `help:"Specify the infra/sandbox image to use" default:"${default_infra_image}"`
	VethCidr   string `help:"CIDR to use for veth pairs. dboxed will dynamically allocate 2 IPs from this CIDR per box" default:"1.2.3.0/24"`
}

func (cmd *RunCmd) Run(g *flags.GlobalFlags, logHandler *logs.MultiLogHandler) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	m, err := commandutils.GetMachine(ctx, c, cmd.Machine)
	if err != nil {
		return err
	}

	runMachine := run_machine.RunMachine{
		Debug:      g.Debug,
		WorkDir:    g.WorkDir,
		InfraImage: cmd.InfraImage,
		VethCidr:   cmd.VethCidr,
		Client:     c,
		MachineId:  m.ID,
	}

	return runMachine.Run(ctx, logHandler)
}
