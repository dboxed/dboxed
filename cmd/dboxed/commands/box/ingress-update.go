package box

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type UpdateIngressCmd struct {
	Box         string  `help:"Box ID or name" required:"" arg:""`
	IngressId   string  `help:"Ingress ID" required:"" arg:""`
	Description *string `help:"Description of the ingress"`
	Hostname    *string `help:"Hostname for ingress"`
	PathPrefix  *string `help:"Path prefix"`
	Port        *int    `help:"Container port"`
}

func (cmd *UpdateIngressCmd) Run(g *flags.GlobalFlags) error {
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

	req := models.UpdateBoxIngress{
		Description: cmd.Description,
		Hostname:    cmd.Hostname,
		PathPrefix:  cmd.PathPrefix,
		Port:        cmd.Port,
	}

	ing, err := c2.UpdateBoxIngress(ctx, b.ID, cmd.IngressId, req)
	if err != nil {
		return err
	}

	slog.Info("Updated box ingress", slog.Any("id", ing.ID))

	return nil
}
