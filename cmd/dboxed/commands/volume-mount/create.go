//go:build linux

package volume_mount

import (
	"context"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/volume/volume_serve"
)

type CreateCmd struct {
	Volume string  `help:"Specify volume" required:"" arg:""`
	Box    *string `help:"Specify the box that wants to mount this volume"`

	MountName *string `help:"Override the local mount name. Defaults to the volume ID"`
}

func (cmd *CreateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	v, err := commandutils.GetVolume(ctx, c, cmd.Volume)
	if err != nil {
		return err
	}

	mountName := v.ID
	if cmd.MountName != nil {
		mountName = *cmd.MountName
	}

	dir := filepath.Join(g.WorkDir, "volumes", v.ID)
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	vsOpts := volume_serve.VolumeServeOpts{
		Client:    c,
		MountName: mountName,
		VolumeId:  v.ID,
		Dir:       dir,
	}

	if cmd.Box != nil {
		b, err := commandutils.GetBox(ctx, c, *cmd.Box)
		if err != nil {
			return err
		}
		vsOpts.BoxId = &b.ID
	}

	vs, err := volume_serve.New(vsOpts)
	if err != nil {
		return err
	}

	err = vs.Create(ctx)
	if err != nil {
		return err
	}

	err = vs.Open(ctx)
	if err != nil {
		return err
	}

	err = vs.MountDevice(ctx, false)
	if err != nil {
		return err
	}

	err = vs.RestoreFromLatestSnapshot(ctx)
	if err != nil {
		return err
	}

	return nil
}
