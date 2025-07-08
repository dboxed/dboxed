package sandbox

import (
	"context"
	"fmt"
	"github.com/koobox/unboxed/pkg/types"
	"github.com/koobox/unboxed/pkg/util"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

func (rn *Sandbox) buildNetbirdContainerSpec() (*types.ContainerSpec, error) {
	if rn.BoxSpec.Netbird.Image == "" && rn.BoxSpec.Netbird.Version == "" {
		return nil, fmt.Errorf("one of image or version must be specified for netbird")
	}

	image := rn.BoxSpec.Netbird.Image
	if image == "" {
		image = fmt.Sprintf("netbirdio/netbird:%s", rn.BoxSpec.Netbird.Version)
	}

	var entrypoint []string
	entrypoint = append(entrypoint, "netbird", "service", "run")

	var env []string
	env = append(env, "NB_FOREGROUND_MODE=false")

	s := &types.ContainerSpec{
		Name:       "_netbird",
		Image:      image,
		Entrypoint: entrypoint,
		Env:        env,
		Privileged: true,
	}
	return s, nil
}

func (rn *Sandbox) runNetbirdUp(ctx context.Context) error {
	setupKeyFile := "/etc/netbird/setup-key"
	setupKeyFileInContainer := filepath.Join(rn.getContainerRoot("_netbird"), setupKeyFile)
	err := os.MkdirAll(filepath.Dir(setupKeyFileInContainer), 0700)
	if err != nil {
		return err
	}
	err = os.WriteFile(setupKeyFileInContainer, []byte(rn.BoxSpec.Netbird.SetupKey), 0600)
	if err != nil {
		return fmt.Errorf("failed to write netbird setup key into container: %w", err)
	}

	var args []string
	args = append(args, "exec", "_netbird")
	args = append(args, "netbird", "up")
	args = append(args, "--management-url", rn.BoxSpec.Netbird.ManagementUrl)
	args = append(args, "--setup-key-file", setupKeyFile)

	_, err = RunRunc(ctx, rn.SandboxDir, false, args...)
	if err != nil {
		return fmt.Errorf("netbird up failed: %w", err)
	}
	return nil
}

func (rn *Sandbox) runNetbirdStatusLoop(ctx context.Context) {
	for {
		err := rn.runNetbirdStatus(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "error in runNetbirdStatusLoop", slog.Any("error", err))
		}
		if !util.SleepWithContext(ctx, 5*time.Second) {
			return
		}
	}
}

func (rn *Sandbox) runNetbirdStatus(ctx context.Context) error {
	statusFileInInfraContainer := filepath.Join(rn.getContainerRoot("_infra"), types.NetbirdStatusFile)

	var args []string
	args = append(args, "exec", "_netbird")
	args = append(args, "netbird", "status", "--json")

	s, err := RunRunc(ctx, rn.SandboxDir, true, args...)
	if err != nil {
		return fmt.Errorf("netbird status failed: %w", err)
	}
	err = os.WriteFile(statusFileInInfraContainer+".tmp", []byte(s), 0600)
	if err != nil {
		return fmt.Errorf("failed to write netbird status into container: %w", err)
	}
	err = os.Rename(statusFileInInfraContainer+".tmp", statusFileInInfraContainer)
	if err != nil {
		return fmt.Errorf("failed to rename netbird status tmp file: %w", err)
	}
	return nil
}
