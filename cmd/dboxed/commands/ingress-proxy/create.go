package ingress_proxy

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type CreateCmd struct {
	Name      string `help:"Proxy name" required:""`
	ProxyType string `help:"Proxy type" default:"caddy" enum:"caddy"`
	Network   string `help:"Attach proxy to specified network (ID or name)." required:"true"`
	HttpPort  int    `help:"HTTP port (default: 80)" default:"80"`
	HttpsPort int    `help:"HTTPS port (default: 443)" default:"443"`
}

func (cmd *CreateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	network, err := commandutils.GetNetwork(ctx, c, cmd.Network)
	if err != nil {
		return err
	}

	c2 := &clients.IngressProxyClient{Client: c}

	req := models.CreateIngressProxy{
		Name:      cmd.Name,
		ProxyType: global.IngressProxyType(cmd.ProxyType),
		Network:   network.ID,
		HttpPort:  cmd.HttpPort,
		HttpsPort: cmd.HttpsPort,
	}

	proxy, err := c2.CreateIngressProxy(ctx, req)
	if err != nil {
		return err
	}

	slog.Info("Created ingress proxy", slog.Any("id", proxy.ID), slog.Any("name", proxy.Name))

	return nil
}
