package service

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/volume-mount/service/service_files"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/service"
	"github.com/dboxed/dboxed/pkg/volume/volume_serve"
)

type InstallCmd struct {
	flags.VolumeServeArgs
}

func (cmd *InstallCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	initSystem, err := service.DetectInitSystem(ctx)
	if err != nil {
		return err
	}
	slog.Info("detected init system", slog.Any("initSystem", initSystem))

	backupInterval, err := time.ParseDuration(cmd.BackupInterval)
	if err != nil {
		return err
	}

	baseDir := filepath.Join(g.WorkDir, "volumes")
	volumeState, err := commandutils.GetMountedVolume(baseDir, cmd.Volume)
	if err != nil {
		return err
	}

	switch initSystem {
	case service.InitSystemS6:
		err = cmd.installS6(ctx, g, volumeState, backupInterval)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("init system %s not supported", initSystem)
	}

	return nil
}

func (cmd *InstallCmd) installS6(ctx context.Context, g *flags.GlobalFlags, volumeState *volume_serve.VolumeState, backupInterval time.Duration) error {
	serviceName := fmt.Sprintf("dboxed-volume-%s", volumeState.MountName)

	scripts, err := service_files.GetS6RunScripts(g.WorkDir, volumeState.MountName, volumeState.Volume.ID, backupInterval)
	if err != nil {
		return err
	}

	s6s := service.S6Service{
		ServiceName:   serviceName,
		RunContent:    scripts.Run,
		RunLogContent: scripts.RunLog,
	}

	err = s6s.Install(ctx)
	if err != nil {
		return err
	}

	err = s6s.Enable(ctx)
	if err != nil {
		return err
	}

	err = s6s.Start(ctx)
	if err != nil {
		return err
	}

	return nil
}
