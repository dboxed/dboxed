package network

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type DeleteCmd struct {
	Network string `help:"Network ID or name" required:"" arg:""`
}

func (cmd *DeleteCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	n, err := commandutils.GetNetwork(ctx, c, cmd.Network)
	if err != nil {
		return err
	}

	c2 := &clients.NetworkClient{Client: c}

	err = c2.DeleteNetwork(ctx, n.ID)
	if err != nil {
		return err
	}

	slog.Info("network deleted", slog.Any("id", n.ID), slog.Any("name", n.Name))

	return nil
}
