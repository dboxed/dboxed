package volume

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type AttachCmd struct {
	Box      string  `help:"Box ID, UUID, or name" required:"" arg:""`
	Volume   string  `help:"Volume ID, UUID, or name" required:"" arg:""`
	RootUid  *int64  `help:"Root UID for volume mount"`
	RootGid  *int64  `help:"Root GID for volume mount"`
	RootMode *string `help:"Root mode for volume mount (octal)"`
}

func (cmd *AttachCmd) Run(g *flags.GlobalFlags) error {
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

	req := models.AttachVolumeRequest{
		VolumeId: v.ID,
		RootUid:  cmd.RootUid,
		RootGid:  cmd.RootGid,
		RootMode: cmd.RootMode,
	}

	err = c2.AttachVolume(ctx, b.ID, req)
	if err != nil {
		return err
	}

	slog.Info("volume attached", slog.Any("box_id", b.ID), slog.Any("volume_id", v.ID), slog.Any("volume_name", v.Name))

	return nil
}
