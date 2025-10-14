//go:build linux

package volume_mount

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/volume/volume_serve"
)

type ServeCmd struct {
	Dir string `help:"Specify the local directory for the volume" required:"" type:"existingdir"`

	BackupInterval string `help:"Specify the backup interval" default:"5m"`

	flags.WebdavProxyFlags
}

func (cmd *ServeCmd) Run() error {
	ctx := context.Background()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	backupInterval, err := time.ParseDuration(cmd.BackupInterval)
	if err != nil {
		return err
	}

	volumeState, err := volume_serve.LoadVolumeState(cmd.Dir)
	if err != nil {
		return err
	}

	vs, err := volume_serve.New(volume_serve.VolumeServeOpts{
		VolumeId:          volumeState.VolumeId,
		BoxUuid:           volumeState.BoxUuid,
		Dir:               cmd.Dir,
		BackupInterval:    backupInterval,
		WebdavProxyListen: cmd.WebdavProxyListen,
	})
	if err != nil {
		return err
	}

	err = vs.Open(ctx)
	if err != nil {
		return err
	}

	err = vs.Lock(ctx)
	if err != nil {
		return err
	}

	err = vs.Mount(ctx, false)
	if err != nil {
		return err
	}

	vs.Start(ctx)

	s := <-sigs

	slog.Info(fmt.Sprintf("received %s, stopping periodic backup", s.String()))
	vs.Stop()

	slog.Info("periodic backup stopped")

	return nil
}
