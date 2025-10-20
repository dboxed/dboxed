package volume

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type ListCmd struct {
	Box string `help:"Box ID, UUID, or name" required:"" arg:""`
}

func (cmd *ListCmd) Run(g *flags.GlobalFlags) error {
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

	attachments, err := c2.ListAttachedVolumes(ctx, b.ID)
	if err != nil {
		return err
	}

	for _, a := range attachments {
		volumeName := ""
		if a.Volume != nil {
			volumeName = a.Volume.Name
		}
		slog.Info("volume attachment",
			slog.Any("volume_id", a.VolumeID),
			slog.Any("volume_name", volumeName),
			slog.Any("root_uid", a.RootUid),
			slog.Any("root_gid", a.RootGid),
			slog.Any("root_mode", a.RootMode))
	}

	return nil
}
