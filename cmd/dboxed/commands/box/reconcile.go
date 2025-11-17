package box

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type ReconcileCmd struct {
	Box string `help:"Specify the box" required:"" arg:""`
}

func (cmd *ReconcileCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.BoxClient{Client: c}

	b, err := commandutils.GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	box, err := c2.ReconcileBox(ctx, b.ID)
	if err != nil {
		return err
	}

	slog.Info("Requested box reconciliation", slog.Any("id", box.ID), slog.Any("name", box.Name))

	return nil
}
