package box

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type AddIngressCmd struct {
	Box         string  `help:"Box ID or name" required:"" arg:""`
	ProxyID     string  `help:"Ingress proxy ID" required:""`
	Hostname    string  `help:"Hostname for ingress" required:""`
	PathPrefix  string  `help:"Path prefix" default:"/"`
	Port        int     `help:"Container port" required:""`
	Description *string `help:"Description of the ingress"`
}

func (cmd *AddIngressCmd) Run(g *flags.GlobalFlags) error {
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

	req := models.CreateBoxIngress{
		ProxyID:     cmd.ProxyID,
		Description: cmd.Description,
		Hostname:    cmd.Hostname,
		PathPrefix:  cmd.PathPrefix,
		Port:        cmd.Port,
	}

	ing, err := c2.CreateBoxIngress(ctx, b.ID, req)
	if err != nil {
		return err
	}

	slog.Info("Created box ingress", slog.Any("id", ing.ID))

	return nil
}
