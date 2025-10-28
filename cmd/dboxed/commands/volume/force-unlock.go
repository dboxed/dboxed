package volume

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type ForceUnlockCmd struct {
	Volume string `help:"Specify volume" required:"" arg:""`
}

func (cmd *ForceUnlockCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	v, err := commandutils.GetVolume(ctx, c, cmd.Volume)
	if err != nil {
		return err
	}

	c2 := clients.VolumesClient{Client: c}

	_, err = c2.VolumeForceUnlock(ctx, v.ID)
	if err != nil {
		return err
	}

	slog.Info("volume force unlocked", slog.Any("id", v.ID), slog.Any("name", v.Name))

	return nil
}
