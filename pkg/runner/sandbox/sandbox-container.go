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

func (rn *Sandbox) KillSandboxContainer(ctx context.Context) error {
	c, err := rn.GetSandboxContainer()
	if err != nil {
		if errors.Is(err, libcontainer.ErrNotExist) {
			return nil
		}
		return err
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
		return err
	}
	if !running {
		return nil
	}

	slog.InfoContext(ctx, "trying to gracefully stop sandbox container")
	err = c.Signal(unix.SIGTERM)
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "waiting for sandbox container to exit")

	running, err = waitRunning(time.Now().Add(time.Second * 10))
	if err != nil {
		return err
	}
	if !running {
		slog.InfoContext(ctx, "sandbox container has exited")
		return nil
	}

	slog.InfoContext(ctx, "sandbox container still running, killing it now")
	err = c.Signal(unix.SIGKILL)
	if err != nil {
		return err
	}

	running, err = waitRunning(time.Now().Add(time.Second * 10))
	if err != nil {
		return err
	}
	if running {
		return fmt.Errorf("failed to stop/kill sandbox container")
	}

	slog.InfoContext(ctx, "sandbox container has exited")
	return nil
}

func (rn *Sandbox) destroySandboxContainer(ctx context.Context) error {
	err := rn.KillSandboxContainer(ctx)
	if err != nil {
		return err
	}

	c, err := rn.GetSandboxContainer()
	if err != nil {
		if errors.Is(err, libcontainer.ErrNotExist) {
			return nil
		}
		return err
	}

	slog.InfoContext(ctx, "destroying old sandbox container")
	err = c.Destroy()
	if err != nil {
		return err
	}

	return nil
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
