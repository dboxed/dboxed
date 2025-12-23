package spec

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type DeleteCmd struct {
	DboxedSpec string `help:"Specify dboxed spec" required:"" arg:""`
}

func (cmd *DeleteCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	gs, err := commandutils.GetDboxedSpec(ctx, c, cmd.DboxedSpec)
	if err != nil {
		return err
	}

	c2 := &clients.DboxedSpecClient{Client: c}

	err = c2.DeleteDboxedSpec(ctx, gs.ID)
	if err != nil {
		return err
	}

	slog.Info("dboxed spec deleted", slog.Any("id", gs.ID))

	return nil
}
