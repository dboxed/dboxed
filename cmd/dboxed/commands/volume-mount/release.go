//go:build linux

package volume_mount

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/volume/volume_serve"
)

type ReleaseCmd struct {
	Dir string `help:"Specify the local directory for the volume" required:"" type:"existingdir"`

	flags.WebdavProxyFlags
}

func (cmd *ReleaseCmd) Run() error {
	ctx := context.Background()

	volumeState, err := volume_serve.LoadVolumeState(cmd.Dir)
	if err != nil {
		return err
	}

	vs, err := volume_serve.New(volume_serve.VolumeServeOpts{
		VolumeId:          volumeState.VolumeId,
		Dir:               cmd.Dir,
		WebdavProxyListen: cmd.WebdavProxyListen,
	})
	if err != nil {
		return err
	}

	err = vs.Open(ctx)
	if err != nil {
		return err
	}

	err = vs.Mount(ctx, true)
	if err != nil {
		return err
	}

	slog.Info("Remounting read-only")
	err = vs.LocalVolume.RemountReadOnly(ctx, vs.GetMountDir())
	if err != nil {
		return err
	}

	slog.Info("performing final backup")
	err = vs.BackupOnce(ctx)
	if err != nil {
		return err
	}

	// we unlock early, because the volume being read-only already ensures we don't lose data
	err = vs.Unlock(ctx)
	if err != nil {
		return err
	}

	err = vs.LocalVolume.Unmount(ctx, vs.GetMountDir())
	if err != nil {
		return err
	}

	err = vs.Deactivate(ctx)
	if err != nil {
		return err
	}

	return nil
}
