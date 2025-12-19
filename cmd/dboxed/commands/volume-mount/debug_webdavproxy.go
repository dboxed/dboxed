package volume_mount

import (
	"context"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/volume/restic_rest_server"
)

type ResticRestServerCmd struct {
	S3Bucket string `name:"s3-bucket" help:"Specify the bucket" required:""`
	S3Prefix string `name:"s3-prefix" help:"Specify the path prefix"`

	ResticRestServerListen string `help:"Specify restic rest server listen address" default:"127.0.0.1:10000"`
}

func (cmd *ResticRestServerCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	b, err := commandutils.GetS3Bucket(ctx, c, cmd.S3Bucket)
	if err != nil {
		return err
	}

	s, err := restic_rest_server.NewServer(c, b.ID, cmd.S3Prefix, cmd.ResticRestServerListen)
	if err != nil {
		return err
	}
	_, err = s.Start(ctx)
	if err != nil {
		return err
	}
	defer s.Stop()

	<-ctx.Done()

	return nil
}
