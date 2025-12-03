package machine

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type RemoveBoxCmd struct {
	Machine string `help:"Machine ID or name" required:"" arg:""`
	Box     string `help:"Box ID or name" required:""`
}

func (cmd *RemoveBoxCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	m, err := commandutils.GetMachine(ctx, c, cmd.Machine)
	if err != nil {
		return err
	}

	b, err := commandutils.GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	c2 := &clients.MachineClient{Client: c}

	err = c2.RemoveBox(ctx, m.ID, b.ID)
	if err != nil {
		return err
	}

	slog.Info("box removed from machine", slog.Any("machine_id", m.ID), slog.Any("machine_name", m.Name), slog.Any("box_id", b.ID), slog.Any("box_name", b.Name))

	return nil
}
