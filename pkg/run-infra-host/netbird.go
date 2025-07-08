package run_infra_host

import (
	"context"
	"fmt"
	"github.com/koobox/unboxed/pkg/sandbox"
	"github.com/koobox/unboxed/pkg/types"
	"github.com/koobox/unboxed/pkg/util"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

func (rn *RunInfraHost) waitForNetbirdContainer(ctx context.Context) error {
	slog.InfoContext(ctx, "waiting for netbird container to start")

	for {
		s, err := sandbox.RunRuncState(ctx, rn.conf.SandboxDir, "_netbird")
		if err != nil {
			return err
		}
		if s.Status == "running" {
			return nil
		}
		if !util.SleepWithContext(ctx, time.Second) {
			return ctx.Err()
		}
	}
}

func (rn *RunInfraHost) runNetbirdCli(ctx context.Context, captureStdout bool, args ...string) (string, error) {
	var args2 []string
	args2 = append(args2, "exec", "_netbird", "netbird")
	args2 = append(args2, args...)
	stdout, err := sandbox.RunRunc(ctx, rn.conf.SandboxDir, captureStdout, args2...)
	if err != nil {
		return stdout, fmt.Errorf("netbird cli failed: %w", err)
	}
	return stdout, nil
}

func (rn *RunInfraHost) runNetbirdUp(ctx context.Context) error {
	slog.InfoContext(ctx, "running netbird up")

	setupKeyFile := filepath.Join(types.SharedDir, "netbird-setup-key")

	err := os.WriteFile(setupKeyFile, []byte(rn.conf.BoxSpec.Netbird.SetupKey), 0600)
	if err != nil {
		return fmt.Errorf("failed to write netbird setup key into container: %w", err)
	}

	_, err = rn.runNetbirdCli(ctx, false, "up",
		"--management-url", rn.conf.BoxSpec.Netbird.ManagementUrl,
		"--setup-key-file", setupKeyFile,
	)
	return err
}

func (rn *RunInfraHost) runNetbirdStatusLoop(ctx context.Context) {
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

func (rn *RunInfraHost) runNetbirdStatus(ctx context.Context) error {
	s, err := rn.runNetbirdCli(ctx, true, "status", "--json")
	if err != nil {
		return err
	}

	err = util.AtomicWriteFile(types.NetbirdStatusFile, []byte(s), 0600)
	if err != nil {
		return fmt.Errorf("failed to write netbird status into shared dir: %w", err)
	}

	return nil
}
