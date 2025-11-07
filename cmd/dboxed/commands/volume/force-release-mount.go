package volume

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type ForceReleaseMountCmd struct {
	Volume string `help:"Specify volume" required:"" arg:""`
}

func (cmd *ForceReleaseMountCmd) Run(g *flags.GlobalFlags) error {
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

	_, err = c2.VolumeForceReleaseMount(ctx, v.ID)
	if err != nil {
		return err
	}

	slog.Info("mount force released", slog.Any("id", v.ID), slog.Any("name", v.Name))

	return nil
}
