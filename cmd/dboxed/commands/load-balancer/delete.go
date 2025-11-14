package load_balancer

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type DeleteCmd struct {
	LoadBalancer string `help:"Specify the load balancer" required:"" arg:""`
}

func (cmd *DeleteCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.LoadBalancerClient{Client: c}

	lb, err := commandutils.GetLoadBalancer(ctx, c, cmd.LoadBalancer)
	if err != nil {
		return err
	}

	err = c2.DeleteLoadBalancer(ctx, lb.ID)
	if err != nil {
		return err
	}

	slog.Info("Deleted load balancer", slog.Any("id", lb.ID))

	return nil
}
