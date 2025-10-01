//go:build linux

package volume

import (
	"context"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/box"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/volume/volume_serve"
)

type LockCmd struct {
	Volume string  `help:"Specify volume" required:"" args:""`
	Box    *string `help:"Specify the box that wants to lock this volume"`

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

	vsOpts := volume_serve.VolumeServeOpts{
		Client:   c,
		VolumeId: v.ID,
		Dir:      cmd.Dir,
	}

	if cmd.Box != nil {
		b, err := box.GetBox(ctx, c, *cmd.Box)
		if err != nil {
			return err
		}
		vsOpts.BoxUuid = &b.Uuid
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
