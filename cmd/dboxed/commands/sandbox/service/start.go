//go:build linux

package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	run_sandbox "github.com/dboxed/dboxed/pkg/runner/run-sandbox"
	"github.com/dboxed/dboxed/pkg/runner/service"
)

type StartCmd struct {
	flags.SandboxArgsRequired
}

func (cmd *StartCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	sandboxBaseDir := run_sandbox.GetSandboxDir(g.WorkDir, "")
	si, err := commandutils.GetSandboxInfo(sandboxBaseDir, &cmd.Sandbox)
	if err != nil {
		return err
	}

	initSystem, err := service.DetectInitSystem(ctx)
	if err != nil {
		return err
	}
	slog.Info("detected init system", slog.Any("initSystem", initSystem))

	switch initSystem {
	case service.InitSystemSystemd:
		unitName := fmt.Sprintf("dboxed-sandbox-%s", si.SandboxName)

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
