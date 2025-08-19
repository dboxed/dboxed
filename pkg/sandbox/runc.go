package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/types"
)

func BuildRuncCmd(ctx context.Context, sandboxDir string, args ...string) *exec.Cmd {
	binPath := filepath.Join(sandboxDir, "runc")
	stateDir := getRuncStateDir(sandboxDir)

	args2 := []string{
		"--root", stateDir,
	}
	args2 = append(args2, args...)

	//slog.InfoContext(ctx, "runc: "+strings.Join(args2, " "))

	cmd := exec.CommandContext(ctx, binPath, args2...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func RunRunc(ctx context.Context, sandboxDir string, captureStdout bool, args ...string) (string, error) {
	cmd := BuildRuncCmd(ctx, sandboxDir, args...)

	stdout := bytes.NewBuffer(nil)
	if captureStdout {
		cmd.Stdout = stdout
	} else {
		cmd.Stdout = os.Stdout
	}

	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return stdout.String(), nil
}

func RunRuncJson(ctx context.Context, sandboxDir string, result any, args ...string) error {
	stdout, err := RunRunc(ctx, sandboxDir, true, args...)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(stdout), result)
	if err != nil {
		return err
	}
	return nil
}

func RunRuncList(ctx context.Context, sandboxDir string) ([]types.RuncState, error) {
	var ret []types.RuncState
	err := RunRuncJson(ctx, sandboxDir, &ret, "list", "--format=json")
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func RunRuncState(ctx context.Context, sandboxDir string, name string) (*types.RuncState, error) {
	var ret types.RuncState
	err := RunRuncJson(ctx, sandboxDir, &ret, "state", name)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}
