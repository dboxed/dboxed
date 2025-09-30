//go:build linux

package volume

import (
	"context"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/volume/volume_serve"
)

type LockCmd struct {
	Volume string `help:"Specify volume" required:"" args:""`

	Dir string `help:"Specify the local directory for the volume" required:"" type:"path"`
}

func (cmd *LockCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient()
	if err != nil {
		return err
	}

	v, err := getVolume(ctx, c, cmd.Volume)
	if err != nil {
		return err
	}

	vs, err := volume_serve.New(volume_serve.VolumeServeOpts{
		Client:   c,
		VolumeId: v.ID,
		Dir:      cmd.Dir,
	})
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

	err = vs.Mount(ctx, false)
	if err != nil {
		return err
	}

	err = vs.RestoreFromLatestSnapshot(ctx)
	if err != nil {
		return err
	}

	return nil
}
