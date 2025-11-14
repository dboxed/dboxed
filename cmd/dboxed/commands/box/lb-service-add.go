package box

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type AddLbServiceCmd struct {
	Box          string  `help:"Box ID or name" required:"" arg:""`
	LoadBalancer string  `help:"Load balancer ID or name" required:""`
	Hostname     string  `help:"Hostname for load balancer service" required:""`
	PathPrefix   string  `help:"Path prefix" default:"/"`
	Port         int     `help:"Sandbox port" required:""`
	Description  *string `help:"Description of the load balancer service"`
}

func (cmd *AddLbServiceCmd) Run(g *flags.GlobalFlags) error {
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

	lb, err := commandutils.GetLoadBalancer(ctx, c, cmd.LoadBalancer)
	if err != nil {
		return err
	}

	req := models.CreateLoadBalancerService{
		LoadBalancerID: lb.ID,
		Description:    cmd.Description,
		Hostname:       cmd.Hostname,
		PathPrefix:     cmd.PathPrefix,
		Port:           cmd.Port,
	}

	ing, err := c2.CreateLoadBalancerService(ctx, b.ID, req)
	if err != nil {
		return err
	}

	slog.Info("Created load balancer service", slog.Any("id", ing.ID))

	return nil
}
