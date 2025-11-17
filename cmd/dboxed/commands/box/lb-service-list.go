package box

import (
	"context"
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type ListLbServicesCmd struct {
	Box string `help:"Box ID or name" required:"" arg:""`
	flags.ListFlags
}

type PrintLoadBalancerService struct {
	ID           string `col:"ID"` // do not set id:true here as the user will actually need the ID to update/remove the port-forward
	Description  string `col:"Description"`
	LoadBalancer string `col:"Load Balancer"`
	Hostname     string `col:"Hostname"`
	PathPrefix   string `col:"Path Prefix"`
	Port         int    `col:"Port"`
}

func (cmd *ListLbServicesCmd) Run(g *flags.GlobalFlags) error {
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
	ct := commandutils.NewClientTool(c)

	lbs, err := c2.ListLoadBalancerServices(ctx, b.ID)
	if err != nil {
		return err
	}

	var table []PrintLoadBalancerService
	for _, lb := range lbs {
		desc := ""
		if lb.Description != nil {
			desc = *lb.Description
		}

		table = append(table, PrintLoadBalancerService{
			ID:           lb.ID,
			Description:  desc,
			LoadBalancer: ct.LoadBalancers.GetColumn(ctx, lb.LoadBalancerID, cmd.ShowIds),
			Hostname:     lb.Hostname,
			PathPrefix:   lb.PathPrefix,
			Port:         lb.Port,
		})
	}

	err = commandutils.PrintTable(os.Stdout, table, cmd.ShowIds)
	if err != nil {
		return err
	}

	return nil
}
