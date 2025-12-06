package box

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type StartCmd struct {
	Box string `help:"Specify the box" required:"" arg:""`
}

func (cmd *StartCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.BoxClient{Client: c}

	b, err := commandutils.GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	updatedBox, err := c2.StartBox(ctx, b.ID)
	if err != nil {
		return err
	}

	slog.Info("box starting",
		slog.Any("id", updatedBox.ID),
		slog.Any("name", updatedBox.Name),
		slog.Any("enabled", updatedBox.Enabled),
	)

	return nil
}
