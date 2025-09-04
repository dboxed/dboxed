package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os/exec"
	"slices"
	"strings"

	"github.com/dboxed/dboxed/pkg/logs/line_handler"
)

func (rn *Sandbox) BuildSandboxCmd(ctx context.Context, workDir string, command string, args ...string) *exec.Cmd {
	var args2 []string
	args2 = append(args2, "exec")
	if workDir != "" {
		args2 = append(args2, "--cwd", workDir)
	}
	args2 = append(args2, "sandbox", command)
	args2 = append(args2, args...)
	cmd := BuildRuncCmd(ctx, rn.SandboxDir, args2...)

	return cmd
}

func (rn *Sandbox) RunSandboxCmd(ctx context.Context, log *slog.Logger, captureStdout bool, workDir string, command string, args ...string) (string, error) {
	stdoutBuf := bytes.NewBuffer(nil)

	var lhStdout, lhStderr io.WriteCloser
	defer func() {
		if lhStdout != nil {
			_ = lhStdout.Close()
		}
		if lhStderr != nil {
			_ = lhStderr.Close()
		}
	}()

	cmd := rn.BuildSandboxCmd(ctx, workDir, command, args...)
	if captureStdout {
		cmd.Stdout = stdoutBuf
	} else {
		lhStdout = line_handler.NewLineHandler(func(line string) {
			log.InfoContext(ctx, line)
		})
		cmd.Stdout = lhStdout
	}

	lhStderr = line_handler.NewLineHandler(func(line string) {
		log.WarnContext(ctx, line)
	})
	cmd.Stderr = lhStderr

	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return stdoutBuf.String(), nil
}

func (rn *Sandbox) BuildDockerCliCmd(ctx context.Context, workDir string, args ...string) *exec.Cmd {
	return rn.BuildSandboxCmd(ctx, workDir, "docker", args...)
}

func (rn *Sandbox) RunDockerCli(ctx context.Context, log *slog.Logger, captureStdout bool, workDir string, args ...string) (string, error) {
	return rn.RunSandboxCmd(ctx, log, captureStdout, workDir, "docker", args...)
}

func (rn *Sandbox) RunDockerCliJson(ctx context.Context, log *slog.Logger, ret any, workDir string, args ...string) error {
	stdout, err := rn.RunDockerCli(ctx, log, true, workDir, args...)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(stdout), ret)
	if err != nil {
		return err
	}
	return nil
}

func (rn *Sandbox) RunDockerCliJsonLines(ctx context.Context, log *slog.Logger, ret any, workDir string, args ...string) error {
	stdout, err := rn.RunDockerCli(ctx, log, true, workDir, args...)
	if err != nil {
		return err
	}
	lines := strings.Split(stdout, "\n")
	lines = slices.DeleteFunc(lines, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})
	array := "[" + strings.Join(lines, ",") + "]"
	err = json.Unmarshal([]byte(array), ret)
	if err != nil {
		return err
	}
	return nil
}
