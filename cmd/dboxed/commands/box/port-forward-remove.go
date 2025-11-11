package box

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type RemovePortForwardCmd struct {
	Box           string `help:"Box ID or name" required:"" arg:""`
	PortForwardId string `help:"Port forward ID" required:"" arg:""`
}

func (cmd *RemovePortForwardCmd) Run(g *flags.GlobalFlags) error {
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

	err = c2.DeletePortForward(ctx, b.ID, cmd.PortForwardId)
	if err != nil {
		return err
	}

	slog.Info("Removed port forward", slog.Any("id", cmd.PortForwardId))

	return nil
}
