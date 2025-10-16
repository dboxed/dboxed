package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

type CommandHelper struct {
	ContainerHolder

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
	if c.LogCmd {
		s := c.Command
		if len(c.Args) != 0 {
			s += strings.Join(c.Args, " ")
		}
		if c.isContainer() {
			slog.InfoContext(ctx, fmt.Sprintf("running command: %s", s))
		} else {
			slog.InfoContext(ctx, fmt.Sprintf("running command in container: %s", s))
		}
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	var stdout, stderr io.Writer
	if c.CatchStdout {
		stdout = &stdoutBuf
	} else {
		stdout = os.Stdout
	}
	if c.CatchStderr {
		stderr = &stderrBuf
	} else {
		stderr = os.Stdout
	}

	var err error
	if c.isContainer() {
		err = c.runContainer(ctx, stdout, stderr)
	} else {
		err = c.runNormal(ctx, stdout, stderr)
	}
	c.Stdout = stdoutBuf.Bytes()
	c.Stderr = stderrBuf.Bytes()
	return err
}

func (c *CommandHelper) runNormal(ctx context.Context, stdout io.Writer, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, c.Command, c.Args...)
	cmd.Dir = c.Dir
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if c.Env != nil {
		cmd.Env = c.Env
	}
	return cmd.Run()
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
