//go:build linux

package volume_mount

import (
	"context"
	"path/filepath"
	"time"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/volume/volume_serve"
)

type MountCmd struct {
	flags.VolumeMountArgs
}

func (cmd *MountCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	_, err := lockAndMountVolume(ctx, g.WorkDir, cmd.Volume, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func lockAndMountVolume(ctx context.Context, workDir string, volume string, backupInterval *time.Duration, webdavProxyListen *string) (*volume_serve.VolumeServe, error) {
	baseDir := filepath.Join(workDir, "volumes")
	volumeState, err := commandutils.GetMountedVolume(baseDir, volume)
	if err != nil {
		return nil, err
	}

	opts := volume_serve.VolumeServeOpts{
		MountName: volumeState.MountName,
		VolumeId:  volumeState.Volume.ID,
		BoxId:     volumeState.BoxId,
		Dir:       filepath.Join(baseDir, volumeState.MountName),
	}
	if backupInterval != nil {
		opts.BackupInterval = *backupInterval
	}
	if webdavProxyListen != nil {
		opts.WebdavProxyListen = *webdavProxyListen
	}

	vs, err := volume_serve.New(opts)
	if err != nil {
		return nil, err
	}

	err = vs.Open(ctx)
	if err != nil {
		return nil, err
	}

	err = vs.Lock(ctx)
	if err != nil {
		return nil, err
	}

	err = vs.Mount(ctx, false)
	if err != nil {
		return nil, err
	}

	return vs, nil
}
