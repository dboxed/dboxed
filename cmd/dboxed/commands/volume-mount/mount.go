//go:build linux

package volume_mount

import (
	"context"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
)

type MountCmd struct {
	Volume string `help:"Specify volume" required:"" arg:""`
}

func (cmd *MountCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	err := runServeVolumeCmd(ctx, g, runServeVolumeCmdOpts{
		volume:  cmd.Volume,
		create:  false,
		mount:   true,
		serve:   false,
		release: false,
	})

	if err != nil {
		return err
	}

	return nil
}
