package volume_mount

import (
	"context"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/volume/webdavproxy"
)

type WebdavProxyCmd struct {
	S3Bucket string `name:"s3-bucket" help:"Specify the bucket" required:""`
	S3Prefix string `name:"s3-prefix" help:"Specify the path prefix"`

	WebdavProxyListen string `help:"Specify Webdav/S3 proxy listen address" default:"127.0.0.1:10000"`
}

func (cmd *WebdavProxyCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	b, err := commandutils.GetS3Bucket(ctx, c, cmd.S3Bucket)
	if err != nil {
		return err
	}

	fs := webdavproxy.NewFileSystem(ctx, c, b.ID, cmd.S3Prefix)

	webdavProxy, err := webdavproxy.NewProxy(fs, cmd.WebdavProxyListen)
	if err != nil {
		return err
	}
	_, err = webdavProxy.Start(ctx)
	if err != nil {
		return err
	}
	defer webdavProxy.Stop()

	<-ctx.Done()

	return nil
}
