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
)

type ServeCmd struct {
	flags.VolumeServeArgs

	flags.WebdavProxyFlags
}

func (cmd *ServeCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	backupInterval, err := time.ParseDuration(cmd.BackupInterval)
	if err != nil {
		return err
	}

	vs, err := lockAndMountVolume(ctx, g.WorkDir, cmd.Volume, &backupInterval, &cmd.WebdavProxyListen)
	if err != nil {
		return err
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

	slog.Info("starting periodic backup", slog.Any("interval", cmd.BackupInterval))
	err = vs.Run(ctx)
	if err != nil {
		return err
	}

	slog.Info("periodic backup stopped")

	return nil
}
