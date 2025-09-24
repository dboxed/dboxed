package volume_provider

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
)

type DeleteCmd struct {
	Volume string `help:"Specify volume provider" required:"" arg:""`
}

func (cmd *DeleteCmd) Run() error {
	ctx := context.Background()

	c, err := baseclient.FromClientAuthFile()
	if err != nil {
		return err
	}

	v, err := GetVolumeProvider(ctx, c, cmd.Volume)
	if err != nil {
		return err
	}

	c2 := clients.VolumeProvidersClient{Client: c}

	err = c2.DeleteVolumeProvider(ctx, v.ID)
	if err != nil {
		return err
	}

	slog.Info("volume provider deleted", slog.Any("id", v.ID))

	return nil
}
