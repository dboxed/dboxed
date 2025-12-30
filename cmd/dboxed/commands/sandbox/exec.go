//go:build linux

package sandbox

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	run_sandbox "github.com/dboxed/dboxed/pkg/runner/run-sandbox"
	"github.com/dboxed/dboxed/pkg/runner/runc_exec"
	"github.com/dboxed/dboxed/pkg/runner/sandbox"
)

type ExecCmd struct {
	flags.SandboxArgsRequired

	Args []string `help:"Args..." arg:""`

	Env  []string `help:"Environment variables"`
	Cwd  string   `help:"Specify working directory"`
	Tty  bool     `help:"Allocate a pseudo-TTY" short:"t"`
	User string   `help:"UID (format: <uid>[:<gid>])"`
}

func (cmd *ExecCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	if len(cmd.Args) < 1 {
		return fmt.Errorf("at least one command argument must be passed")
	}

	sandboxBaseDir := run_sandbox.GetSandboxDir(g.WorkDir, "")
	si, err := commandutils.GetSandboxInfo(sandboxBaseDir, &cmd.Sandbox)
	if err != nil {
		return err
	}

	s := sandbox.Sandbox{
		Debug:       g.Debug,
		HostWorkDir: g.WorkDir,
		SandboxDir:  run_sandbox.GetSandboxDir(g.WorkDir, si.SandboxId),
	}

	c, err := s.GetSandboxContainer()
	if err != nil {
		return err
	}

	imageConfig, err := s.GetInfraImageConfig()
	if err != nil {
		return err
	}

	args := []string{"dummy"}
	args = append(args, cmd.Args...)

	var env []string
	env = append(env, imageConfig.Config.Env...)

	opts := runc_exec.ExecOpts{
		Container: c,
		Args:      args,
		Cwd:       cmd.Cwd,
		Env:       env,
		Tty:       cmd.Tty,
		User:      cmd.User,
	}
	if opts.Cwd == "" {
		opts.Cwd = "/"
	}

	status, err := runc_exec.ExecProcess(opts)
	if err != nil {
		slog.ErrorContext(ctx, "error while running process", slog.Any("error", err))
	}
	os.Exit(status)
	return nil
}
