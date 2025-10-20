package volume

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type DetachCmd struct {
	Box    string `help:"Box ID, UUID, or name" required:"" arg:""`
	Volume string `help:"Volume ID, UUID, or name" required:"" arg:""`
}

func (cmd *DetachCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	b, err := commandutils.GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	v, err := commandutils.GetVolume(ctx, c, cmd.Volume)
	if err != nil {
		return err
	}

	c2 := &clients.BoxClient{Client: c}

	err = c2.DetachVolume(ctx, b.ID, v.ID)
	if err != nil {
		return err
	}

	slog.Info("volume detached", slog.Any("box_id", b.ID), slog.Any("volume_id", v.ID), slog.Any("volume_name", v.Name))

	return nil
}
