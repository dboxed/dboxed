package box

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type RemoveLbServiceCmd struct {
	Box         string `help:"Box ID or name" required:"" arg:""`
	LbServiceId string `help:"Load Balancer Service ID" required:"" arg:""`
}

func (cmd *RemoveLbServiceCmd) Run(g *flags.GlobalFlags) error {
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

	err = c2.DeleteLoadBalancerService(ctx, b.ID, cmd.LbServiceId)
	if err != nil {
		return err
	}

	slog.Info("Removed load balancer service", slog.Any("id", cmd.LbServiceId))

	return nil
}
