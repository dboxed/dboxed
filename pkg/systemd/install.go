package systemd

import (
	"context"
	"fmt"
	"github.com/dboxed/dboxed/pkg/systemd/units"
	"net/url"
	"os"
	"os/exec"
)

type SystemdInstall struct {
	BoxUrl  *url.URL
	Nkey    string
	BoxName string
}

func (s *SystemdInstall) Run(ctx context.Context) error {
	err := os.MkdirAll("/etc/dboxed", 0700)
	if err != nil {
		return err
	}

	envFileContent := fmt.Sprintf("DBOXED_BOX_URL=%s\n", s.BoxUrl)
	err = os.WriteFile("/etc/dboxed/box-url.env", []byte(envFileContent), 0600)
	if err != nil {
		return err
	}

	serviceName := fmt.Sprintf("dboxed-%s", s.BoxName)

	extraArgs := ""
	if s.Nkey != "" {
		extraArgs += " --nkey=" + s.Nkey
	}

	systemdUnitContent := units.GetDboxedUnit(s.BoxName, extraArgs)
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
