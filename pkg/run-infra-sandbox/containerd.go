package run_infra_sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/koobox/unboxed/pkg/logs"
	"log/slog"
	"os/exec"
	"time"
)

func (rn *RunInfraSandbox) startContainerd(ctx context.Context) error {
	slog.InfoContext(ctx, "starting containerd")

	logRot := logs.BuildRotatingLogger("/var/log/unboxed/containerd.log")
	stdout := logs.NewJsonFileLogger(logRot, "stdout")
	stderr := logs.NewJsonFileLogger(logRot, "stderr")
	defer func() {
		_ = stdout.Close()
		_ = stderr.Close()
	}()

	cmd := exec.CommandContext(ctx, "containerd")
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Start()
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "waiting for containerd to become ready")
	time.Sleep(500 * time.Millisecond)
	for {
		_, err = rn.runNerdctl(ctx, false, []string{"ps"})
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	slog.InfoContext(ctx, "containerd became ready")

	return nil
}

func (rn *RunInfraSandbox) runNerdctl(ctx context.Context, captureStdout bool, args ...string) (string, error) {
	stdoutBuf := bytes.NewBuffer(nil)

	cmd := exec.CommandContext(ctx, "nerdctl", args...)
	cmd.Stdout = rn.sandboxStdout
	if captureStdout {
		cmd.Stdout = stdoutBuf
	}
	cmd.Stderr = rn.sandboxStderr
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return stdoutBuf.String(), nil
}

func (rn *RunInfraSandbox) runNerdctlJson(ctx context.Context, ret any, args ...string) error {
	stdout, err := rn.runNerdctl(ctx, true, args...)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(stdout), ret)
	if err != nil {
		return err
	}
	return nil
}
