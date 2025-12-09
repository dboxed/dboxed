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
	flags.MachineRunArgs
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
