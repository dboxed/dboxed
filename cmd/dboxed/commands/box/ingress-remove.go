package box

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type RemoveIngressCmd struct {
	Box       string `help:"Box ID or name" required:"" arg:""`
	IngressId string `help:"Ingress ID" required:"" arg:""`
}

func (cmd *RemoveIngressCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	b, err := commandutils.GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	c2 := &clients.BoxClient{Client: c}

	err = c2.DeleteBoxIngress(ctx, b.ID, cmd.IngressId)
	if err != nil {
		return err
	}

	slog.Info("Removed box ingress", slog.Any("id", cmd.IngressId))

	return nil
}
