package volume_provider

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type CreateCmd struct {
	Name string `help:"Specify the repository name. Must be unique." required:"" arg:""`

	Type string `help:"Specify the provider type." required:"" enum:"rustic"`

	S3Bucket string `name:"s3-bucket" help:"Specify the S3 bucket to use" required:""`

	StoragePrefix  string `help:"Specify the storage prefix"`
	RusticPassword string `help:"Specify the password used for encryption" required:""`
}

func (cmd *CreateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.VolumeProvidersClient{Client: c}

	req := models.CreateVolumeProvider{
		Type: dmodel.VolumeProviderType(cmd.Type),
		Name: cmd.Name,
	}
	req.Rustic = &models.CreateVolumeProviderRustic{
		StorageType:   dmodel.VolumeProviderStorageTypeS3,
		StoragePrefix: cmd.StoragePrefix,
		Password:      cmd.RusticPassword,
	}

	b, err := commandutils.GetS3Bucket(ctx, c, cmd.S3Bucket)
	if err != nil {
		return err
	}

	req.Rustic.S3BucketId = &b.ID

	rep, err := c2.CreateVolumeProvider(ctx, req)
	if err != nil {
		return err
	}

	slog.Info("volume provider created", slog.Any("id", rep.ID))

	return nil
}
