package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/koobox/unboxed/pkg/types"
	"github.com/opencontainers/runtime-spec/specs-go"
	"os"
	"os/exec"
	"path/filepath"
)

func (rn *Sandbox) writeOciSpec(c *types.ContainerSpec, spec *specs.Spec) error {
	pth := filepath.Join(rn.getContainerBundleDir(c.Name), "config.json")

	err := os.MkdirAll(filepath.Dir(pth), 0700)
	if err != nil {
		return err
	}

	b, err := json.Marshal(spec)
	if err != nil {
		return err
	}

	err = os.WriteFile(pth, b, 0600)
	if err != nil {
		return err
	}
	return nil
}

func (rn *Sandbox) copyRuncFromInfraContainer() error {
	infraPth := filepath.Join(rn.getContainerRoot("_infra"), "sbin/runc")
	hostPth := filepath.Join(rn.SandboxDir, "runc")

	r, err := os.ReadFile(infraPth)
	if err != nil {
		return fmt.Errorf("failed to read runc binary from infra container: %w", err)
	}
	err = os.WriteFile(hostPth, r, 0777)
	if err != nil {
		return fmt.Errorf("failed to write runc binary to work dir: %w", err)
	}
	return nil
}

func (rn *Sandbox) runc(ctx context.Context, captureStdout bool, args ...string) (string, error) {
	binPath := filepath.Join(rn.SandboxDir, "runc")
	stateDir := rn.getRuncStateDir()

	err := os.MkdirAll(stateDir, 0700)
	if err != nil {
		return "", err
	}

	stdout := bytes.NewBuffer(nil)

	args2 := []string{
		"--root", stateDir,
	}
	args2 = append(args2, args...)

	cmd := exec.CommandContext(ctx, binPath, args2...)
	if captureStdout {
		cmd.Stdout = stdout
	} else {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return "", err
	}
	return stdout.String(), nil
}

func (rn *Sandbox) runcJson(ctx context.Context, result any, args ...string) error {
	stdout, err := rn.runc(ctx, true, args...)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(stdout), result)
	if err != nil {
		return err
	}
	return nil
}
