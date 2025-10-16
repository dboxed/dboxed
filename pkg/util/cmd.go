package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

type RunCommandOptions struct {
	CatchStdout bool
	Dir         string
}

type CommandHelper struct {
	Command string
	Args    []string
	Env     []string

	CatchStdout bool
	CatchStderr bool
	Dir         string

	Stdout []byte
	Stderr []byte

	LogCmd bool
}

func (c *CommandHelper) Run(ctx context.Context) error {
	var cmdStdout bytes.Buffer
	var cmdStderr bytes.Buffer

	if c.LogCmd {
		s := c.Command
		if len(c.Args) != 0 {
			s += strings.Join(c.Args, " ")
		}
		slog.InfoContext(ctx, fmt.Sprintf("running command: %s", s))
	}

	cmd := exec.CommandContext(ctx, c.Command, c.Args...)
	cmd.Dir = c.Dir
	if c.CatchStdout {
		cmd.Stdout = &cmdStdout
	} else {
		cmd.Stdout = os.Stdout
	}
	if c.CatchStderr {
		cmd.Stderr = &cmdStderr
	} else {
		cmd.Stderr = os.Stdout
	}
	if c.Env != nil {
		cmd.Env = c.Env
	}
	err := cmd.Run()

	c.Stdout = cmdStdout.Bytes()
	c.Stderr = cmdStderr.Bytes()
	return err
}

func (c *CommandHelper) RunStdout(ctx context.Context) ([]byte, error) {
	c.CatchStdout = true
	err := c.Run(ctx)
	return c.Stdout, err
}

func (c *CommandHelper) RunStdoutJson(ctx context.Context, ret any) error {
	stdout, err := c.RunStdout(ctx)
	if err != nil {
		return err
	}
	err = json.Unmarshal(stdout, ret)
	if err != nil {
		return err
	}
	return nil
}

func RunCommand(ctx context.Context, command string, args ...string) error {
	c := CommandHelper{
		Command: command,
		Args:    args,
	}
	return c.Run(ctx)
}

func RunCommandStdout(ctx context.Context, command string, args ...string) ([]byte, error) {
	c := CommandHelper{
		Command:     command,
		Args:        args,
		CatchStdout: true,
	}
	return c.RunStdout(ctx)
}

func RunCommandJson[T any](ctx context.Context, command string, args ...string) (*T, error) {
	var ret T
	c := CommandHelper{
		Command:     command,
		Args:        args,
		CatchStdout: true,
	}
	err := c.RunStdoutJson(ctx, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}
