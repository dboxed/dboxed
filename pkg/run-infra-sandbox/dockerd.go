package run_infra_sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/koobox/unboxed/pkg/logs"
	"log/slog"
	"os"
	"os/exec"
	"time"
)

func (rn *RunInfraSandbox) startDockerd(ctx context.Context) error {
	slog.InfoContext(ctx, "starting dockerd")

	logRot := logs.BuildRotatingLogger("/var/log/unboxed/dockerd.log")
	stdout := logs.NewJsonFileLogger(logRot, "stdout")
	stderr := logs.NewJsonFileLogger(logRot, "stderr")
	defer func() {
		_ = stdout.Close()
		_ = stderr.Close()
	}()

	cmd := exec.CommandContext(ctx, "dockerd-entrypoint.sh", "dockerd", "--host", "unix:///var/run/docker.sock")
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = os.Environ()

	err := cmd.Start()
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "waiting for dockerd to become ready")
	time.Sleep(500 * time.Millisecond)
	for {
		_, err = rn.runDockerCli(ctx, false, "ps")
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	slog.InfoContext(ctx, "dockerd became ready")

	return nil
}

func (rn *RunInfraSandbox) runDockerCli(ctx context.Context, captureStdout bool, args ...string) (string, error) {
	stdoutBuf := bytes.NewBuffer(nil)

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = rn.infraStdout
	if captureStdout {
		cmd.Stdout = stdoutBuf
	}
	cmd.Stderr = rn.infraStderr
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return stdoutBuf.String(), nil
}

func (rn *RunInfraSandbox) runDockerCliJson(ctx context.Context, ret any, args ...string) error {
	stdout, err := rn.runDockerCli(ctx, true, args...)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(stdout), ret)
	if err != nil {
		return err
	}
	return nil
}
