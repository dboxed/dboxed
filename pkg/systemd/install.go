package systemd

import (
	"context"
	"fmt"
	"github.com/koobox/unboxed/pkg/systemd/units"
	"net/url"
	"os"
	"os/exec"
)

type SystemdInstall struct {
	BoxUrl  *url.URL
	BoxName string
}

func (s *SystemdInstall) Run(ctx context.Context) error {
	err := os.MkdirAll("/etc/unboxed", 0700)
	if err != nil {
		return err
	}

	envFileContent := fmt.Sprintf("UNBOXED_BOX_URL=%s\n", s.BoxUrl)
	err = os.WriteFile("/etc/unboxed/box-url.env", []byte(envFileContent), 0600)
	if err != nil {
		return err
	}

	serviceName := fmt.Sprintf("unboxed-%s", s.BoxName)
	systemdUnitContent := units.GetUnboxedUnit(s.BoxName)
	err = os.WriteFile(fmt.Sprintf("/etc/systemd/system/%s.service", serviceName), []byte(systemdUnitContent), 0644)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "systemctl", "daemon-reload")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, "systemctl", "enable", serviceName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, "systemctl", "start", serviceName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
