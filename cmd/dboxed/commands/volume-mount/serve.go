//go:build linux

package volume_mount

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/volume/volume_serve"
)

type ServeCmd struct {
	flags.VolumeServeArgs

	flags.WebdavProxyFlags

	ReadyFile *string `help:"Specify ready marker file"`
}

func (cmd *ServeCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	backupInterval, err := time.ParseDuration(cmd.BackupInterval)
	if err != nil {
		return err
	}

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	volume, err := commandutils.GetVolume(ctx, c, cmd.Volume)
	if err != nil {
		return err
	}

	dir := filepath.Join(g.WorkDir, "volumes", volume.ID)
	vsOpts := volume_serve.VolumeServeOpts{
		Client:            c,
		VolumeId:          volume.ID,
		Dir:               dir,
		BackupInterval:    backupInterval,
		WebdavProxyListen: cmd.WebdavProxyListen,
	}
	if cmd.Box != nil {
		box, err := commandutils.GetBox(ctx, c, *cmd.Box)
		if err != nil {
			return err
		}
		vsOpts.BoxId = &box.ID
	}

	vs, err := volume_serve.New(vsOpts)
	if err != nil {
		return err
	}
	volumeState, err := vs.LoadVolumeState()
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if volumeState == nil {
		err = vs.Create(ctx)
		if err != nil {
			return err
		}
	}
	err = vs.Open(ctx)
	if err != nil {
		return err
	}

	err = vs.MountDevice(ctx, false)
	if err != nil {
		return err
	}

	restoreDone, err := vs.IsRestoreDone()
	if err != nil {
		return err
	}
	if !restoreDone {
		err = vs.RestoreFromLatestSnapshot(ctx)
		if err != nil {
			return err
		}
	}

	slog.Info("performing initial backup")
	err = vs.BackupOnce(ctx)
	if err != nil {
		return err
	}

	go func() {
		s := <-sigs
		slog.Info(fmt.Sprintf("received %s, stopping periodic backup", s.String()))
		vs.Stop()
	}()

	if cmd.ReadyFile != nil {
		err = os.WriteFile(*cmd.ReadyFile, nil, 0644)
		if err != nil {
			return err
		}
	}

	slog.Info("starting periodic backup", slog.Any("interval", cmd.BackupInterval))
	err = vs.Run(ctx)
	if err != nil {
		return err
	}

	slog.Info("periodic backup stopped")

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

	// we release early, because the volume being read-only already ensures we don't lose data
	slog.Info("releasing volume mount")
	err = vs.ReleaseVolumeMountViaApi(ctx)
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
