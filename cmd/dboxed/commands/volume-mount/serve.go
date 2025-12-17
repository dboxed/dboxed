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

	err := runServeVolumeCmd(ctx, g, runServeVolumeCmdOpts{
		volume:            cmd.Volume,
		backupInterval:    &cmd.BackupInterval,
		webdavProxyListen: &cmd.WebdavProxyListen,
		box:               cmd.Box,
		readyFile:         cmd.ReadyFile,
		create:            true,
		mount:             true,
		serve:             true,
		release:           true,
	})
	if err != nil {
		return err
	}
	return nil
}

type runServeVolumeCmdOpts struct {
	volume            string
	backupInterval    *string
	webdavProxyListen *string
	box               *string

	readyFile *string

	create  bool
	mount   bool
	serve   bool
	release bool
}

func runServeVolumeCmd(ctx context.Context, g *flags.GlobalFlags, opts runServeVolumeCmdOpts) error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(sigs)
	}()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	volume, err := commandutils.GetVolume(ctx, c, opts.volume)
	if err != nil {
		return err
	}

	dir := filepath.Join(g.WorkDir, "volumes", volume.ID)
	vsOpts := volume_serve.VolumeServeOpts{
		Client:   c,
		VolumeId: volume.ID,
		Dir:      dir,
	}
	if opts.backupInterval != nil {
		backupInterval, err := time.ParseDuration(*opts.backupInterval)
		if err != nil {
			return err
		}
		vsOpts.BackupInterval = backupInterval
	}
	if opts.webdavProxyListen != nil {
		vsOpts.WebdavProxyListen = *opts.webdavProxyListen
	}

	if opts.box != nil {
		box, err := commandutils.GetBox(ctx, c, *opts.box)
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
		if !opts.create {
			return fmt.Errorf("volume-mount not found")
		}
		err = vs.Create(ctx)
		if err != nil {
			return err
		}
	}

	err = vs.Open(ctx)
	if err != nil {
		return err
	}

	didInitialBackup := false
	if opts.mount {
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

		slog.Info("performing single backup")
		err = vs.BackupOnce(ctx)
		if err != nil {
			return err
		}
		didInitialBackup = true
	}

	if opts.serve {
		go func() {
			s := <-sigs
			slog.Info(fmt.Sprintf("received %s, stopping periodic backup", s.String()))
			vs.Stop()
		}()
	}

	if opts.readyFile != nil {
		err = os.WriteFile(*opts.readyFile, nil, 0644)
		if err != nil {
			return err
		}
	}

	if opts.serve {
		slog.Info("starting periodic backup", slog.Any("interval", opts.backupInterval))
		err = vs.Run(ctx)
		if err != nil {
			return err
		}

		slog.Info("periodic backup stopped")
	}

	if opts.release {
		isMounted, err := vs.LocalVolume.IsMounted()
		if err != nil {
			return err
		}
		if isMounted {
			slog.Info("remounting read-only")
			err = vs.LocalVolume.RemountReadOnly(ctx, vs.GetMountDir())
			if err != nil {
				return err
			}
		} else {
			didRestore, err := vs.IsRestoreDone()
			if err != nil {
				return err
			}
			if didRestore {
				slog.Info("need to mount volume for release to perform final backup")
				err = vs.MountDevice(ctx, true)
				if err != nil {
					return err
				}
				isMounted = true
			}
		}

		if isMounted && (opts.serve || !didInitialBackup) {
			slog.Info("performing final backup")
			err = vs.BackupOnce(ctx)
			if err != nil {
				return err
			}
		}

		// we release early, because the volume being read-only already ensures we don't lose data
		slog.Info("releasing volume mount")
		err = vs.ReleaseVolumeMountViaApi(ctx)
		if err != nil {
			return err
		}

		if isMounted {
			slog.Info("unmounting volume")
			err = vs.LocalVolume.Unmount(ctx, vs.GetMountDir())
			if err != nil {
				return err
			}
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
	}

	return nil
}
