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
	"slices"
	"strings"

	"github.com/dboxed/dboxed/pkg/runner/logs/line_handler"
)

type CommandHelper struct {
	ContainerHolder

	Command string
	Args    []string
	Env     []string

	CatchStdout bool
	CatchStderr bool
	Logger      *slog.Logger
	Dir         string

	Stdout []byte
	Stderr []byte

	LogCmd bool
}

func (c *CommandHelper) Run(ctx context.Context) error {
	if c.LogCmd {
		s := c.Command
		if len(c.Args) != 0 {
			s += " " + strings.Join(c.Args, " ")
		}
		log := c.Logger
		if log == nil {
			log = slog.Default()
		}
		if c.isContainer() {
			log.InfoContext(ctx, fmt.Sprintf("running command in container: %s", s))
		} else {
			log.InfoContext(ctx, fmt.Sprintf("running command: %s", s))
		}
	}

	var lhStdout, lhStderr io.WriteCloser
	if c.Logger != nil {
		lhStdout = line_handler.NewLineHandler(func(line string) {
			c.Logger.InfoContext(ctx, line)
		})
		lhStderr = line_handler.NewLineHandler(func(line string) {
			c.Logger.WarnContext(ctx, line)
		})
		defer lhStdout.Close()
		defer lhStderr.Close()
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	var stdout, stderr io.Writer
	if c.CatchStdout {
		stdout = &stdoutBuf
	} else {
		if c.Logger != nil {
			stdout = lhStdout
		} else {
			stdout = os.Stdout
		}
	}
	if c.CatchStderr {
		stderr = &stderrBuf
	} else {
		if c.Logger != nil {
			stderr = lhStderr
		} else {
			stderr = os.Stderr
		}
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

func (c *CommandHelper) RunStdoutJsonLines(ctx context.Context, ret any) error {
	stdout, err := c.RunStdout(ctx)
	if err != nil {
		return err
	}
	lines := strings.Split(string(stdout), "\n")
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

func RunCommandJsonLines[T any](ctx context.Context, command string, args ...string) ([]T, error) {
	var ret []T
	c := CommandHelper{
		Command:     command,
		Args:        args,
		CatchStdout: true,
	}
	err := c.RunStdoutJsonLines(ctx, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
