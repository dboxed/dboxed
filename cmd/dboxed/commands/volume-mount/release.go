//go:build linux

package volume_mount

import (
	"context"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
)

type ReleaseCmd struct {
	Volume string `help:"Specify volume" required:"" arg:""`

	flags.WebdavProxyFlags
}

func (cmd *ReleaseCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	err := runServeVolumeCmd(ctx, g, runServeVolumeCmdOpts{
		volume:            cmd.Volume,
		webdavProxyListen: &cmd.WebdavProxyListen,
		create:            false,
		serve:             false,
		release:           true,
	})

	if err != nil {
		return err
	}

	return nil
}
