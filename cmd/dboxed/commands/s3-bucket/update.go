package s3_bucket

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type UpdateCmd struct {
	S3Bucket string `help:"S3 bucket ID to update" required:"" arg:""`

	Endpoint        *string `help:"S3 endpoint URL" optional:""`
	Bucket          *string `help:"S3 bucket name" optional:""`
	AccessKeyId     *string `help:"S3 access key ID (must be provided with secret)" optional:""`
	SecretAccessKey *string `help:"S3 secret access key (must be provided with access key)" optional:""`
}

func (cmd *UpdateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	s, err := commandutils.GetS3Bucket(ctx, c, cmd.S3Bucket)
	if err != nil {
		return err
	}

	c2 := clients.S3BucketsClient{Client: c}

	req := models.UpdateS3Bucket{
		Endpoint:        cmd.Endpoint,
		Bucket:          cmd.Bucket,
		AccessKeyId:     cmd.AccessKeyId,
		SecretAccessKey: cmd.SecretAccessKey,
	}

	rep, err := c2.UpdateS3Bucket(ctx, s.ID, req)
	if err != nil {
		return err
	}

	slog.Info("s3 bucket configuration updated", slog.Any("id", rep.ID))

	return nil
}
