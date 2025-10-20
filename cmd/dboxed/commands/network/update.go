package network

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type UpdateCmd struct {
	Network string `help:"Network ID or name" required:"" arg:""`

	// Netbird specific flags
	NetbirdVersion     *string `help:"Netbird version"`
	NetbirdAccessToken *string `help:"Netbird API access token"`
}

func (cmd *UpdateCmd) Run(g *flags.GlobalFlags) error {
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

	req := models.UpdateNetwork{}

	// Only set Netbird if any Netbird field is provided
	if cmd.NetbirdVersion != nil || cmd.NetbirdAccessToken != nil {
		req.Netbird = &models.UpdateNetworkNetbird{
			NetbirdVersion: cmd.NetbirdVersion,
			ApiAccessToken: cmd.NetbirdAccessToken,
		}
	}

	updated, err := c2.UpdateNetwork(ctx, n.ID, req)
	if err != nil {
		return err
	}

	slog.Info("network updated", slog.Any("id", updated.ID), slog.Any("name", updated.Name))

	return nil
}
