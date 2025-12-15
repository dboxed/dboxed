package machine_provider

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type DeleteCmd struct {
	MachineProvider string `help:"Specify machine provider ID or name" required:"" arg:""`
}

func (cmd *DeleteCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	mp, err := commandutils.GetMachineProvider(ctx, c, cmd.MachineProvider)
	if err != nil {
		return err
	}

	c2 := &clients.MachineProviderClient{Client: c}

	err = c2.DeleteMachineProvider(ctx, mp.ID)
	if err != nil {
		return err
	}

	slog.Info("machine provider deleted", slog.Any("id", mp.ID), slog.Any("name", mp.Name))

	return nil
}
