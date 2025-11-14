package box

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type UpdateLbServiceCmd struct {
	Box         string  `help:"Box ID or name" required:"" arg:""`
	LbServiceId string  `help:"Load Balancer Service ID" required:"" arg:""`
	Description *string `help:"Description of the load balancer service"`
	Hostname    *string `help:"Hostname for load balancer service"`
	PathPrefix  *string `help:"Path prefix"`
	Port        *int    `help:"Container port"`
}

func (cmd *UpdateLbServiceCmd) Run(g *flags.GlobalFlags) error {
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

	req := models.UpdateLoadBalancerService{
		Description: cmd.Description,
		Hostname:    cmd.Hostname,
		PathPrefix:  cmd.PathPrefix,
		Port:        cmd.Port,
	}

	ing, err := c2.UpdateLoadBalancerService(ctx, b.ID, cmd.LbServiceId, req)
	if err != nil {
		return err
	}

	slog.Info("Updated load balancer service", slog.Any("id", ing.ID))

	return nil
}
