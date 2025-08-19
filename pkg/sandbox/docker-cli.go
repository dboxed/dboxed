package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"slices"
	"strings"
)

func (rn *Sandbox) BuildDockerCliCmd(ctx context.Context, workDir string, args ...string) *exec.Cmd {
	var args2 []string
	args2 = append(args2, "exec")
	if workDir != "" {
		args2 = append(args2, "--cwd", workDir)
	}
	args2 = append(args2, "sandbox", "docker")
	args2 = append(args2, args...)
	cmd := BuildRuncCmd(ctx, rn.SandboxDir, args2...)

	return cmd
}

func (rn *Sandbox) RunDockerCli(ctx context.Context, captureStdout bool, workDir string, args ...string) (string, error) {
	stdoutBuf := bytes.NewBuffer(nil)

	cmd := rn.BuildDockerCliCmd(ctx, workDir, args...)
	if captureStdout {
		cmd.Stdout = stdoutBuf
	}
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return stdoutBuf.String(), nil
}

func (rn *Sandbox) RunDockerCliJson(ctx context.Context, ret any, workDir string, args ...string) error {
	stdout, err := rn.RunDockerCli(ctx, true, workDir, args...)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(stdout), ret)
	if err != nil {
		return err
	}
	return nil
}

func (rn *Sandbox) RunDockerCliJsonLines(ctx context.Context, ret any, workDir string, args ...string) error {
	stdout, err := rn.RunDockerCli(ctx, true, workDir, args...)
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
