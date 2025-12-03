package machine

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type AddBoxCmd struct {
	Machine string `help:"Machine ID or name" required:"" arg:""`
	Box     string `help:"Box ID or name" required:""`
}

func (cmd *AddBoxCmd) Run(g *flags.GlobalFlags) error {
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

	req := models.AddBoxToMachineRequest{
		BoxId: b.ID,
	}

	err = c2.AddBox(ctx, m.ID, req)
	if err != nil {
		return err
	}

	slog.Info("box added to machine", slog.Any("machine_id", m.ID), slog.Any("machine_name", m.Name), slog.Any("box_id", b.ID), slog.Any("box_name", b.Name))

	return nil
}
