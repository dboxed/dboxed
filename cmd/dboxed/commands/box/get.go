package box

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/baseclient"
)

type GetCmd struct {
	Box string `help:"Specify the box" required:"" arg:""`
}

func (cmd *GetCmd) Run() error {
	ctx := context.Background()

	c, err := baseclient.FromClientAuthFile()
	if err != nil {
		return err
	}

	b, err := GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	slog.Info("box", slog.Any("id", b.ID), slog.Any("name", b.Name), slog.Any("created_at", b.CreatedAt))

	return nil
}
