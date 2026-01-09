//go:build linux

package volume_mount

import (
	"context"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/services"
)

type ReleaseCmd struct {
	Volume string `help:"Specify volume" required:"" arg:""`

	flags.WebdavProxyFlags
}

func (cmd *ReleaseCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	s := &services.VolumesService{Client: c}

	err = s.RunServeVolumeCmd(ctx, g.WorkDir, services.RunServeVolumeCmdOpts{
		Volume:            cmd.Volume,
		WebdavProxyListen: &cmd.WebdavProxyListen,
		Create:            false,
		Mount:             false,
		Serve:             false,
		Release:           true,
	})

	if err != nil {
		return err
	}

	return nil
}
