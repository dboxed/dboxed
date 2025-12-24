//go:build linux

package volume_mount

import (
	"context"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/services"
	"github.com/gin-gonic/gin"
)

type ServeCmd struct {
	flags.VolumeServeArgs

	flags.WebdavProxyFlags

	ReadyFile *string `help:"Specify ready marker file"`
}

func (cmd *ServeCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	// restic rest server is using gin
	gin.SetMode(gin.ReleaseMode)

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	s := &services.VolumesService{Client: c}

	err = s.RunServeVolumeCmd(ctx, g.WorkDir, services.RunServeVolumeCmdOpts{
		Volume:            cmd.Volume,
		BackupInterval:    &cmd.BackupInterval,
		WebdavProxyListen: &cmd.WebdavProxyListen,
		Box:               cmd.Box,
		ReadyFile:         cmd.ReadyFile,
		Create:            true,
		Mount:             true,
		Serve:             true,
		Release:           true,
	})

	if err != nil {
		return err
	}
	return nil
}
