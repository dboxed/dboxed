package volume_provider

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type CreateCmd struct {
	Name string `help:"Specify the repository name. Must be unique." required:"" arg:""`

	S3Endpoint        string  `name:"s3-endpoint" help:"Specify S3 endpoint" default:"https://s3.amazonaws.com"`
	S3Region          *string `name:"s3-region" help:"Specify S3 region" optional:""`
	S3Bucket          string  `name:"s3-bucket" help:"Specify S3 bucket" required:""`
	S3AccessKeyId     string  `name:"s3-access-key-id" help:"Specify S3 access key id" required:""`
	S3SecretAccessKey string  `name:"s3-secret-access-key" help:"Specify S3 secret access key" required:""`

	S3Prefix string `name:"s3-prefix" help:"Specify the s3 prefix"`

	RusticPassword string `help:"Specify the password used for encryption" required:""`
}

func (cmd *CreateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient()
	if err != nil {
		return err
	}

	c2 := &clients.VolumeProvidersClient{Client: c}

	req := models.CreateVolumeProvider{
		Name: cmd.Name,
	}
	req.Rustic = &models.CreateVolumeProviderRustic{
		Password: cmd.RusticPassword,
	}

	req.Rustic.StorageS3 = &models.CreateVolumeProviderStorageS3{
		Endpoint:        cmd.S3Endpoint,
		Region:          cmd.S3Region,
		Bucket:          cmd.S3Bucket,
		AccessKeyId:     cmd.S3AccessKeyId,
		SecretAccessKey: cmd.S3SecretAccessKey,
		Prefix:          cmd.S3Prefix,
	}

	rep, err := c2.CreateVolumeProvider(ctx, req)
	if err != nil {
		return err
	}

	slog.Info("volume provider created", slog.Any("id", rep.ID))

	return nil
}
