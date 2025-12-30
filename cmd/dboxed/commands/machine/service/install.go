//go:build linux

package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/machine/service/service_files"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/service"
	"github.com/dboxed/dboxed/pkg/util"
)

type InstallCmd struct {
	flags.MachineRunArgs
}

func (cmd *InstallCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	machine, err := commandutils.GetMachine(ctx, c, cmd.Machine)
	if err != nil {
		return err
	}

	initSystem, err := service.DetectInitSystem(ctx)
	if err != nil {
		return err
	}
	slog.Info("detected init system", slog.Any("initSystem", initSystem))

	machineDir := filepath.Join(g.WorkDir, "machine")
	err = os.MkdirAll(machineDir, 0700)
	if err != nil {
		return err
	}
	authFile := filepath.Join(machineDir, "service-auth.yaml")
	err = util.AtomicWriteFileYaml(authFile, c.GetClientAuth(true), 0600)
	if err != nil {
		return err
	}

	var extraArgs []string
	extraArgs = append(extraArgs, fmt.Sprintf("--infra-image=%s", cmd.InfraImage))
	extraArgs = append(extraArgs, fmt.Sprintf("--veth-cidr=%s", cmd.VethCidr))
	for i := range extraArgs {
		extraArgs[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(extraArgs[i], "'", "\\'"))
	}

	switch initSystem {
	case service.InitSystemSystemd:
		unitName := "dboxed-machine"
		unitContent := service_files.GetMachineUnit(
			machine.Workspace, machine.ID, authFile,
			strings.Join(extraArgs, " "),
		)

		s := service.SystemdUnit{
			UnitName:    unitName,
			UnitContent: unitContent,
		}
		err = s.Install(ctx)
		if err != nil {
			return err
		}
		err = s.Enable(ctx)
		if err != nil {
			return err
		}
		err = s.Start(ctx)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("init system %s not supported", initSystem)
	}

	return nil
}
