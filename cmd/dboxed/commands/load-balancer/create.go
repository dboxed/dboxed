package load_balancer

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
	Name      string `help:"Load balancer name" required:""`
	Type      string `help:"Load balancer type" default:"caddy" enum:"caddy"`
	Network   string `help:"Attach load balancer to specified network (ID or name)." required:"true"`
	HttpPort  int    `help:"HTTP port" default:"80"`
	HttpsPort int    `help:"HTTPS port" default:"443"`
	Replicas  int    `help:"Number of replicas" default:"1"`
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

	c2 := &clients.LoadBalancerClient{Client: c}

	req := models.CreateLoadBalancer{
		Name:             cmd.Name,
		LoadBalancerType: global.LoadBalancerType(cmd.Type),
		Network:          network.ID,
		HttpPort:         cmd.HttpPort,
		HttpsPort:        cmd.HttpsPort,
		Replicas:         cmd.Replicas,
	}

	lb, err := c2.CreateLoadBalancer(ctx, req)
	if err != nil {
		return err
	}

	slog.Info("Created load balancer", slog.Any("id", lb.ID), slog.Any("name", lb.Name))

	return nil
}
