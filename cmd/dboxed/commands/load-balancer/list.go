package load_balancer

import (
	"context"
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type ListCmd struct {
	flags.ListFlags
}

type PrintLoadBalancer struct {
	ID               string `col:"ID" id:"true"`
	Name             string `col:"Name"`
	Network          string `col:"Network"`
	LoadBalancerType string `col:"Type"`
	Replicas         int    `col:"Replicas"`
	Status           string `col:"Status"`
	StatusDetails    string `col:"Status Detail"`
}

func (cmd *ListCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.LoadBalancerClient{Client: c}
	ct := commandutils.NewClientTool(c)

	proxies, err := c2.ListLoadBalancers(ctx)
	if err != nil {
		return err
	}

	var table []PrintLoadBalancer
	for _, p := range proxies {
		table = append(table, PrintLoadBalancer{
			ID:               p.ID,
			Name:             p.Name,
			Network:          ct.Networks.GetColumn(ctx, p.Network, cmd.ShowIds),
			LoadBalancerType: string(p.LoadBalancerType),
			Replicas:         p.Replicas,
			Status:           p.Status,
			StatusDetails:    p.StatusDetails,
		})
	}

	err = commandutils.PrintTable(os.Stdout, table, cmd.ShowIds)
	if err != nil {
		return err
	}

	return nil
}
