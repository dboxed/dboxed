//go:build linux

package volume_mount

import (
	"context"
	"log/slog"
	"path/filepath"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/volume/volume_serve"
)

type ReleaseCmd struct {
	Volume string `help:"Specify volume" required:"" arg:""`

	flags.WebdavProxyFlags
}

func (cmd *ReleaseCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	baseDir := filepath.Join(g.WorkDir, "volumes")
	volumeState, err := getMountedVolume(baseDir, cmd.Volume)
	if err != nil {
		return err
	}

	vs, err := volume_serve.New(volume_serve.VolumeServeOpts{
		VolumeId:          volumeState.Volume.ID,
		Dir:               filepath.Join(baseDir, volumeState.MountName),
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
