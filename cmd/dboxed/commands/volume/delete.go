package volume

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type DeleteCmd struct {
	Volume string `help:"Specify volume" required:"" arg:""`
}

func (cmd *DeleteCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient()
	if err != nil {
		return err
	}

	v, err := getVolume(ctx, c, cmd.Volume)
	if err != nil {
		return err
	}

	c2 := clients.VolumesClient{Client: c}

	err = c2.DeleteVolume(ctx, v.ID)
	if err != nil {
		return err
	}

	slog.Info("volume deleted", slog.Any("id", v.ID))

	return nil
}
