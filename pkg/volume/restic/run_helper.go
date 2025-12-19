package restic

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/dboxed/dboxed/pkg/util/command_helper"
)

func RunResticCommandJson[T any](ctx context.Context, env []string, args []string) (*T, error) {
	stdout, err := RunResticCommand(ctx, env, true, args)
	if err != nil {
		return nil, err
	}

	var ret T
	err = json.Unmarshal(stdout, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func RunResticCommandStdoutCallback(ctx context.Context, env []string, args []string, stdoutCallback func(line string) bool) error {
	pr, pw := io.Pipe()

	done := make(chan struct{})
	go func() {
		defer close(done)

		s := bufio.NewScanner(pr)
		for s.Scan() {
			line := s.Text()
			if !stdoutCallback(line) {
				_, _ = io.Copy(io.Discard, pr)
				return
			}
		}
	}()

	c := buildResticCommand(env, args)
	c.StdoutStream = pw

	err := c.Run(ctx)

	_ = pw.Close()
	<-done

	if err != nil {
		return err
	}

	return nil
}

func RunResticCommand(ctx context.Context, env []string, catchStdout bool, args []string) ([]byte, error) {
	c := buildResticCommand(env, args)
	c.CatchStdout = catchStdout
	err := c.Run(ctx)
	if err != nil {
		return nil, err
	}

	return c.Stdout, err
}

func buildResticCommand(env []string, args []string) command_helper.CommandHelper {
	env2 := os.Environ()
	env2 = append(env2, env...)

	c := command_helper.CommandHelper{
		Command: "restic",
		Args:    args,
		Env:     env2,
		LogCmd:  true,
	}
	return c
}
