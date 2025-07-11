package sandbox

import (
	"context"
	"fmt"
	"github.com/koobox/unboxed/pkg/types"
	"github.com/koobox/unboxed/pkg/util"
	"github.com/koobox/unboxed/pkg/version"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

func (rn *Sandbox) getInfraContainerDir() string {
	return filepath.Join(rn.SandboxDir, "infra")
}

func (rn *Sandbox) getInfraRoot() string {
	return filepath.Join(rn.getInfraContainerDir(), "rootfs")
}

func (rn *Sandbox) getInfraImageConfig() string {
	return filepath.Join(rn.getInfraContainerDir(), "image-config.json")
}

func getRuncStateDir(sandboxDir string) string {
	return filepath.Join(sandboxDir, "runc-state")
}

func (rn *Sandbox) getInfraImage() string {
	infraImage := rn.BoxSpec.InfraImage
	if infraImage == "" {
		tag := "nightly"
		if !version.IsDummyVersion() {
			tag = version.Version
		}
		infraImage = fmt.Sprintf("%s:%s", types.UnboxedInfraImage, tag)
	}
	return infraImage
}

func (rn *Sandbox) destroyInfraContainer(ctx context.Context) error {
	l, err := RunRuncList(ctx, rn.SandboxDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	var s *types.RuncState
	for _, x := range l {
		if x.Id == "infra" {
			s = &x
			break
		}
	}
	if s == nil {
		return nil
	}

	if s.Status != "stopped" {
		slog.InfoContext(ctx, "killing old infra container")
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

		slog.InfoContext(ctx, "deleting old infra container")
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

	slog.InfoContext(ctx, "removing infra container dir")
	err = os.RemoveAll(rn.getInfraContainerDir())
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (rn *Sandbox) createInfraContainer(ctx context.Context) error {
	slog.InfoContext(ctx, "creating infra container")

	imageConfig, err := util.ReadJsonFile[v1.Image](rn.getInfraImageConfig())
	if err != nil {
		return err
	}

	spec, err := rn.buildInfraContainerOciSpec(imageConfig)
	if err != nil {
		return err
	}
	err = rn.writeInfraContainerOciSpec(spec)
	if err != nil {
		return err
	}

	_, err = RunRunc(ctx, rn.SandboxDir, false, "create", "--bundle", rn.getInfraContainerDir(), "infra")
	if err != nil {
		return err
	}
	return nil
}

func (rn *Sandbox) startInfraContainer(ctx context.Context) error {
	slog.InfoContext(ctx, "starting infra container")

	_, err := RunRunc(ctx, rn.SandboxDir, false, "start", "infra")
	if err != nil {
		return err
	}
	return nil
}

func (rn *Sandbox) copyRuncFromInfraContainer() error {
	infraPth := filepath.Join(rn.getInfraRoot(), "usr/bin/runc")
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

func (rn *Sandbox) copyUnboxedBinIntoInfraRoot() error {
	containerPth := filepath.Join(rn.getInfraRoot(), "bin/unboxed")
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
		return fmt.Errorf("failed to write unboxed binary into infra container: %w", err)
	}
	return nil
}
