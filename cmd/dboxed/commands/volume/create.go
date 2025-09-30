package volume

import (
	"context"
	"fmt"
	"log/slog"

	volume_provider "github.com/dboxed/dboxed/cmd/dboxed/commands/volume-provider"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dustin/go-humanize"
)

type CreateCmd struct {
	Name string `help:"Specify the volume name. Must be unique in the repository." required:"true" arg:""`

	VolumeProvider string `help:"Specify the volume provider" required:""`

	FsType string `help:"Specify the filesystem type" default:"ext4"`
	FsSize string `help:"Specify the maximum filesystem size." required:""`
}

func (cmd *CreateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient()
	if err != nil {
		return err
	}

	c2 := clients.VolumesClient{Client: c}

	fsSize, err := humanize.ParseBytes(cmd.FsSize)
	if err != nil {
		return err
	}

	vp, err := volume_provider.GetVolumeProvider(ctx, c, cmd.VolumeProvider)
	if err != nil {
		return err
	}

	req := models.CreateVolume{
		Name:           cmd.Name,
		VolumeProvider: vp.ID,
	}

	switch vp.Type {
	case dmodel.VolumeProviderTypeRustic:
		req.Rustic = &models.CreateVolumeRustic{
			FsSize: int64(fsSize),
			FsType: cmd.FsType,
		}
	default:
		return fmt.Errorf("unsupported volume provider type %s", vp.Type)
	}

	rep, err := c2.CreateVolume(ctx, req)
	if err != nil {
		return err
	}

	slog.Info("volume created", slog.Any("id", rep.ID), slog.Any("uuid", rep.Uuid))

	return nil
}
