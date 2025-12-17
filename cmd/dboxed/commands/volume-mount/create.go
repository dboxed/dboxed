//go:build linux

package volume_mount

import (
	"context"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
)

type CreateCmd struct {
	Volume string `help:"Specify volume" required:"" arg:""`

	flags.WebdavProxyFlags
}

func (cmd *CreateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	err := runServeVolumeCmd(ctx, g, runServeVolumeCmdOpts{
		volume:            cmd.Volume,
		webdavProxyListen: &cmd.WebdavProxyListen,
		create:            true,
		mount:             false,
		serve:             false,
		release:           false,
	})

	if err != nil {
		return err
	}

	return nil
}
