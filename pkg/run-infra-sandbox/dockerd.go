package run_infra_sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/dboxed/dboxed/pkg/logs"
	"github.com/dboxed/dboxed/pkg/types"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func (rn *RunInfraSandbox) startDockerd(ctx context.Context) error {
	slog.InfoContext(ctx, "starting dockerd")

	stdout := logs.BuildRotatingLogger(filepath.Join(types.LogsDir, "dockerd.stdout.log")) // this should actually never get stuff written to
	stderr := logs.BuildRotatingLogger(filepath.Join(types.LogsDir, "dockerd.log"))        // gets structured json logs written to

	cmd := exec.CommandContext(ctx,
		"dockerd-entrypoint.sh",
		"dockerd", "--host",
		"unix:///var/run/docker.sock",
		"--log-format=json",
	)
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

func (rn *RunInfraSandbox) buildDockerCliCmd(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = rn.dockerCliStdout
	cmd.Stderr = rn.dockerCliStderr
	return cmd
}

func (rn *RunInfraSandbox) runDockerCli(ctx context.Context, captureStdout bool, args ...string) (string, error) {
	stdoutBuf := bytes.NewBuffer(nil)

	cmd := rn.buildDockerCliCmd(ctx, args...)
	if captureStdout {
		cmd.Stdout = stdoutBuf
	}
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
