//go:build linux

package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/service"
)

type StartCmd struct {
	SandboxName string `help:"Specify the local sandbox name" required:"" arg:""`
}

func (cmd *StartCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	initSystem, err := service.DetectInitSystem(ctx)
	if err != nil {
		return err
	}
	slog.Info("detected init system", slog.Any("initSystem", initSystem))

	switch initSystem {
	case service.InitSystemSystemd:
		unitName := fmt.Sprintf("dboxed-sandbox-%s", cmd.SandboxName)

		s := service.SystemdUnit{
			UnitName: unitName,
		}
		err = s.Start(ctx)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("init system %s not suppoert", initSystem)
	}

	return nil
}
