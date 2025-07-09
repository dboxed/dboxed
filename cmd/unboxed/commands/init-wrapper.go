package commands

import (
	"context"
	"github.com/koobox/unboxed/cmd/unboxed/flags"
	"gopkg.in/natefinch/lumberjack.v2"
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

	stdoutFile := os.Getenv("UNBOXED_STDOUT_FILE")
	stderrFile := os.Getenv("UNBOXED_STDERR_FILE")
	_ = os.Unsetenv("UNBOXED_STDOUT_FILE")
	_ = os.Unsetenv("UNBOXED_STDERR_FILE")

	slog.InfoContext(ctx, "in init-wrapper",
		slog.Any("stdoutFile", stdoutFile),
		slog.Any("stderrFile", stderrFile),
	)

	stdout := &lumberjack.Logger{
		Filename:   stdoutFile,
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}
	stderr := &lumberjack.Logger{
		Filename:   stderrFile,
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}

	c := exec.CommandContext(ctx, cmd.Args[0], cmd.Args[1:]...)
	c.Stdout = stdout
	c.Stderr = stderr
	c.Stdin = os.Stdin

	return c.Run()
}
