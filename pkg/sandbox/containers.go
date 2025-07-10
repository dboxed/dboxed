package sandbox

import (
	"context"
	"fmt"
	"github.com/koobox/unboxed/pkg/types"
	"github.com/koobox/unboxed/pkg/util"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/opencontainers/runtime-spec/specs-go"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"time"
)

func (rn *Sandbox) getSharedDirOnHost() string {
	return filepath.Join(rn.SandboxDir, "shared")
}

func (rn *Sandbox) getContainerDir(name string) string {
	return filepath.Join(rn.SandboxDir, "containers", name)
}

func (rn *Sandbox) getContainerLogsDir(name string) string {
	return filepath.Join(rn.getContainerDir(name), "logs")
}

func (rn *Sandbox) getContainerRoot(name string) string {
	return filepath.Join(rn.getContainerDir(name), "rootfs")
}

func (rn *Sandbox) getContainerImageConfig(name string) string {
	return filepath.Join(rn.getContainerDir(name), "image-config.json")
}

func getRuncStateDir(sandboxDir string) string {
	return filepath.Join(sandboxDir, "runc-state")
}

func (rn *Sandbox) destroyContainers(ctx context.Context) error {
	stateDir := getRuncStateDir(rn.SandboxDir)

	if _, err := os.Stat(stateDir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var l []types.RuncState
	err := RunRuncJson(ctx, rn.SandboxDir, &l, "list", "--format=json")
	if err != nil {
		return err
	}

	for _, s := range l {
		err = rn.destroyContainer(ctx, s)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rn *Sandbox) destroyContainer(ctx context.Context, s types.RuncState) error {
	log := slog.With(slog.Any("containerName", s.Id))

	if s.Status != "stopped" {
		log.InfoContext(ctx, "killing old container")
		_, err := RunRunc(ctx, rn.SandboxDir, false, "kill", s.Id)
		if err != nil {
			return err
		}
	}

	startTime := time.Now()
	force := false
	for {
		if time.Now().After(startTime.Add(time.Second * 5)) {
			return fmt.Errorf("timed out while trying to delete container %s", s.Id)
		}

		log.InfoContext(ctx, "deleting old container")
		args := []string{"delete"}
		if force {
			args = append(args, "--force")
		}
		args = append(args, s.Id)
		_, err := RunRunc(ctx, rn.SandboxDir, false, args...)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
		force = true
	}

	log.InfoContext(ctx, "removing container dir")
	err := os.RemoveAll(rn.getContainerDir(s.Id))
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (rn *Sandbox) createContainer(ctx context.Context, c *types.ContainerSpec) error {
	slog.InfoContext(ctx, "creating new container", slog.Any("name", c.Name))

	imageConfig, err := util.ReadJsonFile[v1.Image](rn.getContainerImageConfig(c.Name))
	if err != nil {
		return err
	}

	spec, err := rn.buildOciSpec(c, imageConfig)
	if err != nil {
		return err
	}
	err = rn.writeOciSpec(c, spec)
	if err != nil {
		return err
	}

	err = rn.fixBindMountOwnerships(c, spec)
	if err != nil {
		return err
	}

	_, err = RunRunc(ctx, rn.SandboxDir, false, "create", "--bundle", rn.getContainerDir(c.Name), c.Name)
	if err != nil {
		return err
	}
	return nil
}

func (rn *Sandbox) startContainer(ctx context.Context, c *types.ContainerSpec) error {
	slog.InfoContext(ctx, "starting container", slog.Any("name", c.Name))

	_, err := RunRunc(ctx, rn.SandboxDir, false, "start", c.Name)
	if err != nil {
		return err
	}
	return nil
}

func (rn *Sandbox) fixBindMountOwnerships(c *types.ContainerSpec, spec *specs.Spec) error {
	for _, m := range spec.Mounts {
		if !slices.Contains(m.Options, "bind") && !slices.Contains(m.Options, "rbind") {
			continue
		}

		hostPath := m.Source
		containerPath := filepath.Join(rn.getContainerRoot(c.Name), m.Destination)

		err := os.MkdirAll(hostPath, 0755)
		if err != nil {
			return err
		}
		err = os.MkdirAll(containerPath, 0755)
		if err != nil {
			return err
		}

		err = os.Chown(containerPath, int(spec.Process.User.UID), int(spec.Process.User.GID))
		if err != nil {
			return err
		}
	}

	err := os.Chown(rn.getContainerLogsDir(c.Name), int(spec.Process.User.UID), int(spec.Process.User.GID))
	if err != nil {
		return err
	}

	return nil
}

func (rn *Sandbox) copyUnboxedBinIntoContainer(name string) error {
	containerPth := filepath.Join(rn.getContainerRoot(name), "bin/unboxed")
	hostPth, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine unboxed binary path: %w", err)
	}

	r, err := os.ReadFile(hostPth)
	if err != nil {
		return fmt.Errorf("failed to read unboxed binary from host filesystem: %w", err)
	}
	err = util.AtomicWriteFile(containerPth, r, 0777)
	if err != nil {
		return fmt.Errorf("failed to write unboxed binary into infra container %s: %w", name, err)
	}
	return nil
}
