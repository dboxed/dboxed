package systemd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/runner/systemd/units"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
	"sigs.k8s.io/yaml"
)

type SystemdInstall struct {
	ClientAuth *baseclient.ClientAuth
	Box        *models.Box
	LocalName  string
}

func (s *SystemdInstall) Run(ctx context.Context) error {
	err := os.MkdirAll("/etc/dboxed", 0700)
	if err != nil {
		return err
	}

	if s.ClientAuth.StaticToken == nil {
		return fmt.Errorf("systemd units only work with static tokens")
	}

	clientAuthPath := filepath.Join("/etc/dboxed/", fmt.Sprintf("client-auth-%s.yaml", s.LocalName))
	b, err := yaml.Marshal(s.ClientAuth)
	if err != nil {
		return err
	}
	err = util.AtomicWriteFile(clientAuthPath, b, 0600)
	if err != nil {
		return err
	}

	serviceName := fmt.Sprintf("dboxed-%s", s.LocalName)
	extraArgs := []string{
		fmt.Sprintf("--workspace=%d", s.Box.Workspace),
		fmt.Sprintf("%d", s.Box.ID),
	}

	systemdUnitContent := units.GetDboxedUnit(s.LocalName, clientAuthPath, strings.Join(extraArgs, " "))
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
