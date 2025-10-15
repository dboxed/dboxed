package service

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/service"
	"github.com/dboxed/dboxed/pkg/volume/volume_serve"
)

type UninstallCmd struct {
	Volume string `help:"Specify volume" required:"" arg:""`
}

func (cmd *UninstallCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	initSystem, err := service.DetectInitSystem(ctx)
	if err != nil {
		return err
	}
	slog.Info("detected init system", slog.Any("initSystem", initSystem))

	baseDir := filepath.Join(g.WorkDir, "volumes")
	volumeState, err := commandutils.GetMountedVolume(baseDir, cmd.Volume)
	if err != nil {
		return err
	}

	switch initSystem {
	case service.InitSystemS6:
		err = cmd.uninstallS6(ctx, g, volumeState)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("init system %s not suppoert", initSystem)
	}

	return nil
}

func (cmd *UninstallCmd) uninstallS6(ctx context.Context, g *flags.GlobalFlags, volumeState *volume_serve.VolumeState) error {
	serviceName := fmt.Sprintf("dboxed-volume-%s", volumeState.MountName)

	s6s := service.S6Service{
		ServiceName:   serviceName,
	}

	err := s6s.Stop(ctx)
	if err != nil {
		return err
	}

	err = s6s.Uninstall(ctx)
	if err != nil {
		return err
	}

	return nil
}
