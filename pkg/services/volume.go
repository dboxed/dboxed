package services

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
	"github.com/dboxed/dboxed/pkg/baseclient"

	// "github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/volume/volume_serve"
	// "github.com/gin-gonic/gin"
)

type VolumesService struct {
	Client *baseclient.Client
}

type RunServeVolumeCmdOpts struct {
	Volume            string
	BackupInterval    *string
	WebdavProxyListen *string
	Box               *string

	ReadyFile *string
	Create    bool
	Mount     bool
	Serve     bool
	Release   bool
}

func (service *VolumesService) RunServeVolumeCmd(ctx context.Context, workDir string, opts RunServeVolumeCmdOpts) error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(sigs)
	}()

	volume, err := commandutils.GetVolume(ctx, service.Client, opts.Volume)
	if err != nil {
		return err
	}

	dir := filepath.Join(workDir, "volumes", volume.ID)
	vsOpts := volume_serve.VolumeServeOpts{
		Client:   service.Client,
		VolumeId: volume.ID,
		Dir:      dir,
	}
	if opts.BackupInterval != nil {
		backupInterval, err := time.ParseDuration(*opts.BackupInterval)
		if err != nil {
			return err
		}
		vsOpts.BackupInterval = backupInterval
	}
	if opts.WebdavProxyListen != nil {
		vsOpts.ResticRestServerListen = *opts.WebdavProxyListen
	}

	if opts.Box != nil {
		box, err := commandutils.GetBox(ctx, service.Client, *opts.Box)
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
		if !opts.Create {
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
	if opts.Mount {
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

	if opts.Serve {
		go func() {
			s := <-sigs
			slog.Info(fmt.Sprintf("received %s, stopping periodic backup", s.String()))
			vs.Stop()
		}()
	}

	if opts.ReadyFile != nil {
		err = os.WriteFile(*opts.ReadyFile, nil, 0644)
		if err != nil {
			return err
		}
	}

	if opts.Serve {
		slog.Info("starting periodic backup", slog.Any("interval", opts.BackupInterval))
		err = vs.Run(ctx)
		if err != nil {
			return err
		}

		slog.Info("periodic backup stopped")
	}

	if opts.Release {
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

		if isMounted && (opts.Serve || !didInitialBackup) {
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
