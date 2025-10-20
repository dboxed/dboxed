package network

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type ListCmd struct{}

func (cmd *ListCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.NetworkClient{Client: c}

	networks, err := c2.ListNetworks(ctx)
	if err != nil {
		return err
	}

	for _, n := range networks {
		slog.Info("network", slog.Any("id", n.ID), slog.Any("name", n.Name), slog.Any("type", n.Type), slog.Any("status", n.Status))
	}

	return nil
}
