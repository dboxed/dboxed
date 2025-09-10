package sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/dboxed/dboxed-common/util"
	"github.com/dboxed/dboxed/pkg/types"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func (rn *Sandbox) getSandboxContainerDir() string {
	return filepath.Join(rn.SandboxDir, "runc-sandbox")
}

func (rn *Sandbox) GetSandboxRoot() string {
	return filepath.Join(rn.SandboxDir, "sandbox-rootfs")
}

func (rn *Sandbox) getInfraImageConfig() string {
	return filepath.Join(rn.SandboxDir, "infra-image-config.json")
}

func getRuncStateDir(sandboxDir string) string {
	return filepath.Join(sandboxDir, "runc-state")
}

func (rn *Sandbox) RuncState(ctx context.Context) (*types.RuncState, error) {
	l, err := RunRuncList(ctx, rn.SandboxDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		return nil, nil
	}
	for _, s := range l {
		if s.Id == "sandbox" {
			return &s, nil
		}
	}
	return nil, nil
}

func (rn *Sandbox) killSandboxContainerWithSignal(ctx context.Context, id string, signal string) error {
	slog.InfoContext(ctx, fmt.Sprintf("sending %s signal to container %s", signal, id))
	_, err := RunRunc(ctx, rn.SandboxDir, false, "kill", id, signal)
	if err != nil {
		return err
	}
	return nil
}

func (rn *Sandbox) killSandboxContainer(ctx context.Context) error {
	checkAnyRunning := func() (bool, error) {
		l, err := RunRuncList(ctx, rn.SandboxDir)
		if err != nil {
			return false, err
		}
		for _, s := range l {
			if s.Status == "running" {
				return true, nil
			}
		}
		return false, nil
	}
	waitAnyRunning := func(deadline time.Time) (bool, error) {
		for time.Now().Before(deadline) {
			anyRunning, err := checkAnyRunning()
			if err != nil {
				return false, err
			}
			if !anyRunning {
				return false, nil
			}
			if !util.SleepWithContext(ctx, time.Millisecond*500) {
				return false, ctx.Err()
			}
		}
		return true, nil
	}

	anyRunning, err := checkAnyRunning()
	if err != nil {
		return err
	}
	if !anyRunning {
		return nil
	}

	slog.InfoContext(ctx, "trying to gracefully stop sandbox container")
	l, err := RunRuncList(ctx, rn.SandboxDir)
	if err != nil {
		return err
	}
	for _, s := range l {
		if s.Status == "running" {
			_ = rn.killSandboxContainerWithSignal(ctx, s.Id, "TERM")
		}
	}
	slog.InfoContext(ctx, "waiting for sandbox container to exit")

	anyRunning, err = waitAnyRunning(time.Now().Add(time.Second * 10))
	if err != nil {
		return err
	}
	if !anyRunning {
		slog.InfoContext(ctx, "sandbox container has exited")
		return nil
	}

	slog.InfoContext(ctx, "sandbox container still running, killing it now")
	l, err = RunRuncList(ctx, rn.SandboxDir)
	if err != nil {
		return nil
	}
	for _, s := range l {
		if s.Status == "running" {
			_ = rn.killSandboxContainerWithSignal(ctx, s.Id, "KILL")
		}
	}
	anyRunning, err = waitAnyRunning(time.Now().Add(time.Second * 10))
	if err != nil {
		return err
	}
	if anyRunning {
		return fmt.Errorf("failed to stop/kill sandbox container")
	}

	slog.InfoContext(ctx, "sandbox container has exited")

	return nil
}

func (rn *Sandbox) destroySandboxContainer(ctx context.Context) error {
	err := rn.killSandboxContainer(ctx)
	if err != nil {
		return err
	}

	l, err := RunRuncList(ctx, rn.SandboxDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	for _, s := range l {
		slog.InfoContext(ctx, fmt.Sprintf("deleting old %s container", s.Id))
		_, err := RunRunc(ctx, rn.SandboxDir, false, "delete", s.Id)
		if err != nil {
			return err
		}
	}

	slog.InfoContext(ctx, "removing sandbox container dir")
	err = os.RemoveAll(rn.getSandboxContainerDir())
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (rn *Sandbox) createSandboxContainer(ctx context.Context) error {
	slog.InfoContext(ctx, "creating sandbox container")

	b, err := os.ReadFile(rn.getInfraImageConfig())
	if err != nil {
		return err
	}

	var imageConfig v1.Image
	err = json.Unmarshal(b, &imageConfig)
	if err != nil {
		return err
	}

	err = rn.writeSandboxContainerOciSpec(spec)
	if err != nil {
		return err
	}

	_, err = RunRunc(ctx, rn.SandboxDir, false, "create", "--bundle", rn.getSandboxContainerDir(), "sandbox")
	if err != nil {
		return err
	}
	return nil
}

func (rn *Sandbox) startSandboxContainer(ctx context.Context) error {
	slog.InfoContext(ctx, "starting sandbox container")

	_, err := RunRunc(ctx, rn.SandboxDir, false, "start", "sandbox")
	if err != nil {
		return err
	}
	return nil
}

func (rn *Sandbox) copyRuncFromInfraRoot() error {
	infraPth := filepath.Join(rn.GetSandboxRoot(), "usr/local/bin/runc")
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
