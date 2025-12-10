package compose

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/util"
)

func RunComposeCli(ctx context.Context, log *slog.Logger, dir string, projectName string, cmdEnv []string, catchStd bool, args ...string) ([]byte, []byte, error) {
	var args2 []string
	args2 = append(args2, "compose")
	if projectName != "" {
		args2 = append(args2, "-p", projectName)
	}
	args2 = append(args2, args...)

	if log == nil {
		log = slog.Default()
	}

	if projectName != "" {
		log = log.With("composeProject", projectName)
	}

	cmd := util.CommandHelper{
		Command: "docker",
		Args:    args2,
		Env:     cmdEnv,
		Dir:     dir,
		Logger:  log,
		LogCmd:  true,
	}
	if catchStd {
		cmd.CatchStdout = true
		cmd.CatchStderr = true
	}
	err := cmd.Run(ctx)
	if err != nil {
		return nil, nil, err
	}
	return cmd.Stdout, cmd.Stderr, nil
}
