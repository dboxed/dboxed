package s3_bucket

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type CreateCmd struct {
	Endpoint        string `help:"S3 endpoint URL (e.g., https://s3.amazonaws.com)" required:""`
	Bucket          string `help:"S3 bucket name" required:""`
	AccessKeyId     string `help:"S3 access key ID" required:""`
	SecretAccessKey string `help:"S3 secret access key" required:""`
}

func (cmd *CreateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := clients.S3BucketsClient{Client: c}

	req := models.CreateS3Bucket{
		Endpoint:        cmd.Endpoint,
		Bucket:          cmd.Bucket,
		AccessKeyId:     cmd.AccessKeyId,
		SecretAccessKey: cmd.SecretAccessKey,
	}

	rep, err := c2.CreateS3Bucket(ctx, req)
	if err != nil {
		return err
	}

	slog.Info("s3 bucket configuration created", slog.Any("id", rep.ID))

	return nil
}
