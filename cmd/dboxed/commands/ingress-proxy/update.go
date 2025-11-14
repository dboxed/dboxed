package ingress_proxy

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type UpdateCmd struct {
	Proxy     string `help:"Ingress proxy (ID or name)" required:"" arg:""`
	HttpPort  *int   `help:"HTTP port"`
	HttpsPort *int   `help:"HTTPS port"`
	Replicas  *int   `help:"Number of replicas"`
}

func (cmd *UpdateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	proxy, err := commandutils.GetIngressProxy(ctx, c, cmd.Proxy)
	if err != nil {
		return err
	}

	c2 := &clients.IngressProxyClient{Client: c}

	req := models.UpdateIngressProxy{
		HttpPort:  cmd.HttpPort,
		HttpsPort: cmd.HttpsPort,
		Replicas:  cmd.Replicas,
	}

	proxy, err = c2.UpdateIngressProxy(ctx, proxy.ID, req)
	if err != nil {
		return err
	}

	slog.Info("Updated ingress proxy", slog.Any("id", proxy.ID), slog.Any("name", proxy.Name))

	return nil
}
