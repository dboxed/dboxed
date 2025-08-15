package sandbox

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/dboxed/dboxed/pkg/types"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/dboxed/dboxed/pkg/version"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func (rn *Sandbox) getInfraContainerDir(name string) string {
	return filepath.Join(rn.SandboxDir, name)
}

func (rn *Sandbox) getInfraRoot() string {
	return filepath.Join(rn.SandboxDir, "infra-rootfs")
}

func (rn *Sandbox) getInfraImageConfig() string {
	return filepath.Join(rn.SandboxDir, "infra-image-config.json")
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
		infraImage = fmt.Sprintf("%s:%s", types.DboxedInfraImage, tag)
	}
	return infraImage
}

func (rn *Sandbox) destroyInfraContainer(ctx context.Context, name string) error {
	l, err := RunRuncList(ctx, rn.SandboxDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	var s *types.RuncState
	for _, x := range l {
		if x.Id == name {
			s = &x
			break
		}
	}
	if s == nil {
		return nil
	}

	if s.Status != "stopped" {
		slog.InfoContext(ctx, fmt.Sprintf("killing old %s container", name))
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

		slog.InfoContext(ctx, fmt.Sprintf("deleting old %s container", name))
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

	slog.InfoContext(ctx, fmt.Sprintf("removing %s container dir", name))
	err = os.RemoveAll(rn.getInfraContainerDir(name))
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (rn *Sandbox) createInfraContainer(ctx context.Context, hostNetwork bool, name string, cmd []string) error {
	slog.InfoContext(ctx, fmt.Sprintf("creating %s container", name))

	imageConfig, err := util.ReadJsonFile[v1.Image](rn.getInfraImageConfig())
	if err != nil {
		return err
	}

	spec, err := rn.buildInfraContainerOciSpec(imageConfig, hostNetwork, name, cmd)
	if err != nil {
		return err
	}
	err = rn.writeInfraContainerOciSpec(name, spec)
	if err != nil {
		return err
	}

	_, err = RunRunc(ctx, rn.SandboxDir, false, "create", "--bundle", rn.getInfraContainerDir(name), name)
	if err != nil {
		return err
	}
	return nil
}

func (rn *Sandbox) startInfraContainer(ctx context.Context, name string) error {
	slog.InfoContext(ctx, fmt.Sprintf("starting %s container", name))

	_, err := RunRunc(ctx, rn.SandboxDir, false, "start", name)
	if err != nil {
		return err
	}
	return nil
}

func (rn *Sandbox) copyRuncFromInfraRoot() error {
	infraPth := filepath.Join(rn.getInfraRoot(), "usr/local/bin/runc")
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

func (rn *Sandbox) copyDboxedBinIntoInfraRoot() error {
	containerPth := filepath.Join(rn.getInfraRoot(), "bin/dboxed")
	hostPth, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine dboxed binary path: %w", err)
	}

	r, err := os.ReadFile(hostPth)
	if err != nil {
		return fmt.Errorf("failed to read dboxed binary from host filesystem: %w", err)
	}
	err = util.AtomicWriteFile(containerPth, r, 0777)
	if err != nil {
		return fmt.Errorf("failed to write dboxed binary into infra container: %w", err)
	}
	return nil
}
