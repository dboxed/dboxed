//go:build linux

package sandbox

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/util"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/opencontainers/runc/libcontainer"
	"golang.org/x/sys/unix"
	"sigs.k8s.io/yaml"

	_ "github.com/opencontainers/cgroups/devices"
)

func (rn *Sandbox) GetSandboxRoot() string {
	return filepath.Join(rn.SandboxDir, "sandbox-rootfs")
}

func (rn *Sandbox) getInfraImageConfig() string {
	return filepath.Join(rn.SandboxDir, "infra-image-config.json")
}

func GetContainerStateDir(sandboxDir string) string {
	return filepath.Join(sandboxDir, "sandbox-state")
}

func (rn *Sandbox) GetSandboxContainer() (*libcontainer.Container, error) {
	return libcontainer.Load(GetContainerStateDir(rn.SandboxDir), "sandbox")
}

func (rn *Sandbox) StopSandboxContainer(ctx context.Context, timeout time.Duration) error {
	stopped, err := rn.KillSandboxContainer(ctx, unix.SIGTERM, timeout)
	if err != nil {
		return err
	}
	if !stopped {
		return fmt.Errorf("failed to stop sandbox container")
	}
	return nil
}

func (rn *Sandbox) StopOrKillSandboxContainer(ctx context.Context) error {
	stopped, err := rn.KillSandboxContainer(ctx, unix.SIGTERM, time.Second*10)
	if err != nil {
		return err
	}
	if !stopped {
		stopped, err = rn.KillSandboxContainer(ctx, unix.SIGKILL, time.Second*10)
		if err != nil {
			return err
		}
		if !stopped {
			return fmt.Errorf("failed to stop/kill sandbox container")
		}
	}
	return nil
}

func (rn *Sandbox) KillSandboxContainer(ctx context.Context, signal os.Signal, timeout time.Duration) (bool, error) {
	c, err := rn.GetSandboxContainer()
	if err != nil {
		if errors.Is(err, libcontainer.ErrNotExist) {
			return true, nil
		}
		return false, err
	}

	checkRunning := func() (bool, error) {
		s, err := c.Status()
		if err != nil {
			return false, err
		}
		if s == libcontainer.Running {
			return true, nil
		}
		return false, nil
	}
	waitRunning := func(deadline time.Time) (bool, error) {
		for time.Now().Before(deadline) {
			running, err := checkRunning()
			if err != nil {
				return false, err
			}
			if !running {
				return false, nil
			}
			if !util.SleepWithContext(ctx, time.Millisecond*500) {
				return false, ctx.Err()
			}
		}
		return checkRunning()
	}

	running, err := checkRunning()
	if err != nil {
		return false, err
	}
	if !running {
		return true, nil
	}

	slog.InfoContext(ctx, "sending signal to sandbox container", slog.Any("signal", signal.String()))
	err = c.Signal(signal)
	if err != nil {
		return false, err
	}
	slog.InfoContext(ctx, "waiting for sandbox container to exit")

	running, err = waitRunning(time.Now().Add(timeout))
	if err != nil {
		return false, err
	}
	if !running {
		slog.InfoContext(ctx, "sandbox container has exited")
		return true, nil
	}
	return false, fmt.Errorf("timeout while waiting for sandbox container to exit")
}

func (rn *Sandbox) createAndStartSandboxContainer(ctx context.Context) error {
	slog.InfoContext(ctx, "creating sandbox container")

	imageConfig, err := util.UnmarshalYamlFile[v1.Image](rn.getInfraImageConfig())
	if err != nil {
		return err
	}

	b, err := yaml.Marshal(rn.network.Config)
	if err != nil {
		return err
	}
	err = util.AtomicWriteFile(filepath.Join(rn.GetSandboxRoot(), consts.NetworkConfFile), b, 0644)
	if err != nil {
		return err
	}

	config, err := rn.buildSandboxContainerConfig(imageConfig)
	if err != nil {
		return err
	}

	var c *libcontainer.Container
	c, err = libcontainer.Create(GetContainerStateDir(rn.SandboxDir), "sandbox", config)
	if err != nil {
		return err
	}

	process, err := rn.buildSandboxContainerProcessSpec(imageConfig)
	if err != nil {
		return err
	}

	err = c.Run(process)
	if err != nil {
		return err
	}
	return nil
}

func (rn *Sandbox) copyDboxedIntoInfraRoot() error {
	infraPth := filepath.Join(rn.GetSandboxRoot(), "usr/bin/dboxed")
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	b, err := os.ReadFile(exePath)
	if err != nil {
		return err
	}
	err = util.AtomicWriteFile(infraPth, b, 0777)
	if err != nil {
		return err
	}
	return nil
}
