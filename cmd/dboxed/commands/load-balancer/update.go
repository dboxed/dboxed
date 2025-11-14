package load_balancer

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type UpdateCmd struct {
	LoadBalancer string `help:"Specify the load balancer" required:"" arg:""`
	HttpPort     *int   `help:"HTTP port"`
	HttpsPort    *int   `help:"HTTPS port"`
	Replicas     *int   `help:"Number of replicas"`
}

func (cmd *UpdateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	lb, err := commandutils.GetLoadBalancer(ctx, c, cmd.LoadBalancer)
	if err != nil {
		return err
	}

	c2 := &clients.LoadBalancerClient{Client: c}

	req := models.UpdateLoadBalancer{
		HttpPort:  cmd.HttpPort,
		HttpsPort: cmd.HttpsPort,
		Replicas:  cmd.Replicas,
	}

	lb, err = c2.UpdateLoadBalancer(ctx, lb.ID, req)
	if err != nil {
		return err
	}

	slog.Info("Updated load balancer", slog.Any("id", lb.ID), slog.Any("name", lb.Name))

	return nil
}
