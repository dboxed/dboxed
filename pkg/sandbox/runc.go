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
	infraPth := filepath.Join(rn.getContainerRoot("_infra_sandbox"), "sbin/runc")
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

func BuildRuncCmd(ctx context.Context, sandboxDir string, args ...string) (*exec.Cmd, error) {
	binPath := filepath.Join(sandboxDir, "runc")
	stateDir := getRuncStateDir(sandboxDir)

	err := os.MkdirAll(stateDir, 0700)
	if err != nil {
		return nil, err
	}

	args2 := []string{
		"--root", stateDir,
	}
	args2 = append(args2, args...)

	//slog.InfoContext(ctx, "runc: "+strings.Join(args2, " "))

	cmd := exec.CommandContext(ctx, binPath, args2...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd, nil
}

func RunRunc(ctx context.Context, sandboxDir string, captureStdout bool, args ...string) (string, error) {
	cmd, err := BuildRuncCmd(ctx, sandboxDir, args...)
	if err != nil {
		return "", err
	}

	stdout := bytes.NewBuffer(nil)
	if captureStdout {
		cmd.Stdout = stdout
	} else {
		cmd.Stdout = os.Stdout
	}

	err = cmd.Run()
	if err != nil {
		return "", err
	}
	return stdout.String(), nil
}

func RunRuncJson(ctx context.Context, sandboxDir string, result any, args ...string) error {
	stdout, err := RunRunc(ctx, sandboxDir, true, args...)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(stdout), result)
	if err != nil {
		return err
	}
	return nil
}

func RunRuncState(ctx context.Context, sandboxDir string, name string) (*types.RuncState, error) {
	var ret types.RuncState
	err := RunRuncJson(ctx, sandboxDir, &ret, "state", name)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}
