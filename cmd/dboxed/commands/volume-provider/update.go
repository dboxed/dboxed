package volume_provider

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type UpdateCmd struct {
	VolumeProvider string `help:"Specify the volume provider." required:""`

	S3Bucket *string `name:"s3-bucket" help:"Specify the S3 bucket to use"`

	StoragePrefix  *string `help:"Specify the storage prefix"`
	ResticPassword *string `help:"Specify the password used for encryption"`
}

func (cmd *UpdateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.VolumeProvidersClient{Client: c}

	vp, err := commandutils.GetVolumeProvider(ctx, c, cmd.VolumeProvider)
	if err != nil {
		return err
	}

	req := models.UpdateVolumeProvider{}

	req.Restic = &models.UpdateVolumeProviderRestic{
		Password: cmd.ResticPassword,
	}

	var bucketId *string
	if cmd.S3Bucket != nil {
		b, err := commandutils.GetS3Bucket(ctx, c, *cmd.S3Bucket)
		if err != nil {
			return err
		}
		bucketId = &b.ID
	}

	req.Restic.StorageS3 = &models.UpdateRepositoryStorageS3{
		S3BucketId:    bucketId,
		StoragePrefix: cmd.StoragePrefix,
	}

	rep, err := c2.UpdateVolumeProvider(ctx, vp.ID, req)
	if err != nil {
		return err
	}

	slog.Info("volume provider updated", slog.Any("id", rep.ID))

	return nil
}
