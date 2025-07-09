package commands

import (
	"context"
	"github.com/koobox/unboxed/cmd/unboxed/flags"
	"github.com/koobox/unboxed/pkg/logs"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

type InitWrapperCmd struct {
	Args []string `arg:"" optional:"" passthrough:""`
}

func waitPidLoop() {
	for {
		if wpid, _ := syscall.Wait4(-1, nil, syscall.WNOHANG, nil); wpid <= 0 {
			break
		} else {
			slog.Info("Reaped zombie process: pid=%d", wpid)
		}
	}
}

func (cmd *InitWrapperCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	if os.Getpid() == 1 {
		waitPidLoop()

		// Forward signals to children
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGCHLD)
		go func() {
			for s := range sigs {
				if s == syscall.SIGCHLD {
					waitPidLoop()
				}
			}
		}()
	}

	logFile := os.Getenv("UNBOXED_LOG_FILE")
	_ = os.Unsetenv("UNBOXED_LOG_FILE")

	slog.InfoContext(ctx, "in init-wrapper",
		slog.Any("logFile", logFile),
	)

	logRot := logs.BuildRotatingLogger(logFile)

	stdout := logs.NewJsonFileLogger(logRot, "stdout")
	stderr := logs.NewJsonFileLogger(logRot, "stderr")

	c := exec.CommandContext(ctx, cmd.Args[0], cmd.Args[1:]...)
	c.Stdout = stdout
	c.Stderr = stderr
	c.Stdin = os.Stdin

	err := c.Run()
	_ = stdout.Wait()
	_ = stderr.Wait()
	return err
}
