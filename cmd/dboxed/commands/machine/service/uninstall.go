//go:build linux

package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/service"
)

type UninstallCmd struct {
}

func (cmd *UninstallCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	initSystem, err := service.DetectInitSystem(ctx)
	if err != nil {
		return err
	}
	slog.Info("detected init system", slog.Any("initSystem", initSystem))

	switch initSystem {
	case service.InitSystemSystemd:
		unitName := "dboxed-machine"

		s := service.SystemdUnit{
			UnitName: unitName,
		}
		err = s.Uninstall(ctx)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("init system %s not supported", initSystem)
	}

	return nil
}
