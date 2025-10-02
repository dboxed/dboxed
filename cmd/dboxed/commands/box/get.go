package box

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
)

type GetCmd struct {
	Box string `help:"Specify the box" required:"" arg:""`
}

func (cmd *GetCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	b, err := commandutils.GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	slog.Info("box", slog.Any("id", b.ID), slog.Any("name", b.Name), slog.Any("created_at", b.CreatedAt))

	return nil
}
