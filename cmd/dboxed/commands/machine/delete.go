package machine

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type DeleteCmd struct {
	Machine string `help:"Specify the machine ID or name" required:"" arg:""`
}

func (cmd *DeleteCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	m, err := commandutils.GetMachine(ctx, c, cmd.Machine)
	if err != nil {
		return err
	}

	c2 := &clients.MachineClient{Client: c}

	err = c2.DeleteMachine(ctx, m.ID)
	if err != nil {
		return err
	}

	slog.Info("machine deleted", slog.Any("id", m.ID), slog.Any("name", m.Name))

	return nil
}
