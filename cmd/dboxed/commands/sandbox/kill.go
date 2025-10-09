//go:build linux

package sandbox

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/sandbox"
)

type KillCmd struct {
	SandboxName *string `help:"Specify the local sandbox name" optional:"" arg:""`

	All    bool    `help:"Kill all running sandboxes"`
	Signal *string `help:"Specify the signal to be sent to the init process"`
}

func (cmd *KillCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	sandboxBaseDir := filepath.Join(g.WorkDir, "sandboxes")

	var killSandboxes []sandbox.SandboxInfo
	if cmd.SandboxName != nil {
		si, err := sandbox.ReadSandboxInfo(filepath.Join(sandboxBaseDir, *cmd.SandboxName))
		if err != nil {
			return err
		}
		killSandboxes = append(killSandboxes, *si)
	} else if cmd.All {
		var err error
		killSandboxes, err = sandbox.ListSandboxes(sandboxBaseDir)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("you must either specify a sandbox name or use --all")
	}

	for _, si := range killSandboxes {
		sandboxDir := filepath.Join(g.WorkDir, "sandboxes", si.SandboxName)

		args := []string{
			"kill",
			"sandbox",
		}
		if cmd.Signal != nil {
			args = append(args, *cmd.Signal)
		}

		c := sandbox.BuildRuncCmd(ctx, sandboxDir, args...)
		err := c.Run()
		if err != nil {
			slog.Error("runc kill failed", slog.Any("sandboxName", si.SandboxName), slog.Any("error", err))
		}
	}
	return nil
}
