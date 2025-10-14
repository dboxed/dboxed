package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/dboxed/dboxed/pkg/util"
)

type SystemdUnit struct {
	UnitName    string
	UnitContent string
}

func (s *SystemdUnit) Install(ctx context.Context) error {
	unitPath := fmt.Sprintf("/etc/systemd/system/%s.service", s.UnitName)
	err := util.AtomicWriteFile(unitPath, []byte(s.UnitContent), 0644)
	if err != nil {
		return err
	}

	err = SystemdDaemonReload(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *SystemdUnit) Uninstall(ctx context.Context) error {
	err := s.Stop(ctx)
	if err != nil {
		return err
	}
	unitPath := fmt.Sprintf("/etc/systemd/system/%s.service", s.UnitName)
	err = os.Remove(unitPath)
	if err != nil {
		return err
	}
	return nil
}

func (s *SystemdUnit) Enable(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "systemctl", "enable", s.UnitName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (s *SystemdUnit) Start(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "systemctl", "start", s.UnitName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (s *SystemdUnit) Stop(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "systemctl", "stop", s.UnitName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func SystemdDaemonReload(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "systemctl", "daemon-reload")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
