//go:build linux

package sandbox

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/opencontainers/runc/libcontainer"
)

type Sandbox struct {
	Debug bool

	HostWorkDir string

	InfraImage string
	SandboxId  string
	SandboxDir string

	NetworkNamespaceName string
}

func (rn *Sandbox) Destroy(ctx context.Context) error {
	c, err := rn.GetSandboxContainer()
	if err != nil {
		if !errors.Is(err, libcontainer.ErrNotExist) && !os.IsNotExist(err) {
			return err
		}
	} else {
		err = rn.StopOrKillSandboxContainer(ctx, time.Second*30, time.Second*10)
		if err != nil {
			return err
		}
		slog.InfoContext(ctx, "destroying old sandbox container")
		err = c.Destroy()
		if err != nil {
			return err
		}
	}

	slog.InfoContext(ctx, "removing sandbox rootfs")
	err = os.RemoveAll(rn.GetSandboxRoot())
	if err != nil {
		return err
	}
	return nil
}

func (rn *Sandbox) Prepare(ctx context.Context) error {
	err := os.MkdirAll(GetContainerStateDir(rn.SandboxDir), 0700)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Join(rn.SandboxDir, "logs"), 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(rn.SandboxDir, "netbird"), 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(rn.SandboxDir, "volumes"), 0700)
	if err != nil {
		return err
	}

	err = rn.pullInfraImage(ctx)
	if err != nil {
		return err
	}

	return nil

}

func (rn *Sandbox) CopyBinaries(ctx context.Context) error {
	err := rn.copyDboxedIntoInfraRoot()
	if err != nil {
		return err
	}

	return nil
}

func (rn *Sandbox) Start(ctx context.Context) error {
	err := rn.createAndStartSandboxContainer(ctx)
	if err != nil {
		return err
	}

	return nil
}
