package ingress_proxy

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type DeleteCmd struct {
	Id string `help:"Ingress proxy ID" required:"" arg:""`
}

func (cmd *DeleteCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.IngressProxyClient{Client: c}

	err = c2.DeleteIngressProxy(ctx, cmd.Id)
	if err != nil {
		return err
	}

	slog.Info("Deleted ingress proxy", slog.Any("id", cmd.Id))

	return nil
}
