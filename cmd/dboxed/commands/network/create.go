package network

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type CreateCmd struct {
	Name string `help:"Specify the network name. Must be unique." required:"" arg:""`

	Type string `help:"Specify network type. Currently only netbird is supported" enum:"netbird" required:""`

	// Netbird specific flags
	NetbirdVersion     string  `help:"Netbird version" default:"0.59.7"`
	NetbirdApiUrl      *string `help:"Netbird API URL"`
	NetbirdAccessToken string  `help:"Netbird API access token" required:""`
}

func (cmd *CreateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.NetworkClient{Client: c}

	req := models.CreateNetwork{
		Name: cmd.Name,
		Type: global.NetworkNetbird,
	}

	switch cmd.Type {
	case "netbird":
		req.Netbird = &models.CreateNetworkNetbird{
			NetbirdVersion: cmd.NetbirdVersion,
			ApiUrl:         cmd.NetbirdApiUrl,
			ApiAccessToken: cmd.NetbirdAccessToken,
		}
	default:
		return fmt.Errorf("unsupported network type %s", cmd.Type)
	}

	n, err := c2.CreateNetwork(ctx, req)
	if err != nil {
		return err
	}

	slog.Info("network created", slog.Any("id", n.ID), slog.Any("name", n.Name), slog.Any("type", n.Type))

	return nil
}
