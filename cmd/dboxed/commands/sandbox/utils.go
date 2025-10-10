//go:build linux

package sandbox

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/runner/sandbox"
)

func getOneOrAllSandboxes(sandboxBaseDir string, sandboxName *string, all bool) ([]sandbox.SandboxInfo, error) {
	var sandboxes []sandbox.SandboxInfo
	if sandboxName != nil {
		si, err := sandbox.ReadSandboxInfo(filepath.Join(sandboxBaseDir, *sandboxName))
		if err != nil {
			return nil, err
		}
		sandboxes = append(sandboxes, *si)
	} else if all {
		var err error
		sandboxes, err = sandbox.ListSandboxes(sandboxBaseDir)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("you must either specify a sandbox name or use --all")
	}
	return sandboxes, nil
}

func killSandbox(ctx context.Context, sandboxDir string, signal *string) error {
	args := []string{
		"kill",
		"sandbox",
	}
	if signal != nil {
		args = append(args, *signal)
	}

	c := sandbox.BuildRuncCmd(ctx, sandboxDir, args...)
	err := c.Run()
	if err != nil {
		return err
	}

	return nil
}
