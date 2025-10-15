//go:build linux

package volume_mount

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
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
	volumeState, err := commandutils.GetMountedVolume(baseDir, cmd.Volume)
	if err != nil {
		return err
	}
	dir := filepath.Join(baseDir, volumeState.MountName)

	vs, err := volume_serve.New(volume_serve.VolumeServeOpts{
		VolumeId:          volumeState.Volume.ID,
		Dir:               dir,
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

	slog.Info("remounting read-only")
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
	slog.Info("releasing volume lock")
	err = vs.Unlock(ctx)
	if err != nil {
		return err
	}

	slog.Info("unmounting volume")
	err = vs.LocalVolume.Unmount(ctx, vs.GetMountDir())
	if err != nil {
		return err
	}

	err = vs.Deactivate(ctx)
	if err != nil {
		return err
	}

	slog.Info("removing volume dir")
	err = os.RemoveAll(dir)
	if err != nil {
		return err
	}

	return nil
}
