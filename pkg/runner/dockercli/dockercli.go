package dockercli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os/exec"
	"slices"
	"strings"

	"github.com/dboxed/dboxed/pkg/runner/logs/line_handler"
)

func BuildCmd(ctx context.Context, workDir string, command string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = workDir
	return cmd
}

func RunCmd(ctx context.Context, log *slog.Logger, captureStdout bool, workDir string, command string, args ...string) (string, error) {
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

	cmd := BuildCmd(ctx, workDir, command, args...)
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

func BuildDockerCliCmd(ctx context.Context, workDir string, args ...string) *exec.Cmd {
	return BuildCmd(ctx, workDir, "docker", args...)
}

func RunDockerCli(ctx context.Context, log *slog.Logger, captureStdout bool, workDir string, args ...string) (string, error) {
	return RunCmd(ctx, log, captureStdout, workDir, "docker", args...)
}

func RunDockerCliJson(ctx context.Context, log *slog.Logger, ret any, workDir string, args ...string) error {
	stdout, err := RunDockerCli(ctx, log, true, workDir, args...)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(stdout), ret)
	if err != nil {
		return err
	}
	return nil
}

func RunDockerCliJsonLines(ctx context.Context, log *slog.Logger, ret any, workDir string, args ...string) error {
	stdout, err := RunDockerCli(ctx, log, true, workDir, args...)
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
