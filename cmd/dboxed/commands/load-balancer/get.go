package load_balancer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
)

type GetCmd struct {
	LoadBalancer string `help:"Specify the load balancer" required:"" arg:""`
}

func (cmd *GetCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}
	
	lb, err := commandutils.GetLoadBalancer(ctx, c, cmd.LoadBalancer)
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(lb, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(b))

	return nil
}
