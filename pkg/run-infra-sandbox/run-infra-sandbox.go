package run_infra_sandbox

import (
	"context"
	"github.com/koobox/unboxed/pkg/logs"
	"github.com/koobox/unboxed/pkg/sandbox"
	"github.com/koobox/unboxed/pkg/types"
	"github.com/koobox/unboxed/pkg/util"
	"io"
	"log/slog"
	"os"
	"time"
)

type RunInfraSandbox struct {
	conf *types.InfraConfig

	infraStdout io.WriteCloser
	infraStderr io.WriteCloser
}

func (rn *RunInfraSandbox) Run(ctx context.Context) {
	rn.initLogging()
	defer rn.stopLogging()

	err := rn.doRun(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "run-infra-sandbox failed", slog.Any("error", err))
		time.Sleep(time.Minute * 60)
		os.Exit(1)
	}
	os.Exit(0)
}

func (rn *RunInfraSandbox) doRun(ctx context.Context) error {
	slog.InfoContext(ctx, "running in sandbox container")

	var err error
	rn.conf, err = sandbox.ReadInfraConf(types.InfraConfFile)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "waiting for infra-host to become ready")
	for {
		if _, err := os.Stat(types.InfraHostReadyMarkerFile); err == nil {
			break
		}
		if !util.SleepWithContext(ctx, time.Second) {
			return ctx.Err()
		}
	}

	err = rn.startDockerd(ctx)
	if err != nil {
		return err
	}

	err = rn.createBundleVolumes(ctx)
	if err != nil {
		return err
	}

	err = rn.runComposeUp(ctx)
	if err != nil {
		return err
	}

	// let the GC free it up
	rn.conf.BoxSpec.FileBundles = nil

	slog.InfoContext(ctx, "up and running")
	for {
		if !util.SleepWithContext(ctx, 1*time.Second) {
			break
		}
	}

	return nil
}

func (rn *RunInfraSandbox) initLogging() {
	infraLog := logs.BuildRotatingLogger("/var/log/unboxed/infra-sandbox.log")
	rn.infraStdout = logs.NewJsonFileLogger(infraLog, "stdout")
	rn.infraStderr = logs.NewJsonFileLogger(infraLog, "stderr")

	handler := slog.NewTextHandler(rn.infraStderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))
}

func (rn *RunInfraSandbox) stopLogging() {
	_ = rn.infraStderr.Close()
	_ = rn.infraStdout.Close()
}
