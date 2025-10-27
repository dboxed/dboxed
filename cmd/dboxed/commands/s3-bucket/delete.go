package s3_bucket

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type DeleteCmd struct {
	S3Bucket string `help:"S3 bucket ID to delete" required:"" arg:""`
}

func (cmd *DeleteCmd) Run(g *flags.GlobalFlags) error {
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

	err = c2.DeleteS3Bucket(ctx, s.ID)
	if err != nil {
		return err
	}

	slog.Info("s3 bucket configuration deleted", slog.Any("id", s.ID))

	return nil
}
