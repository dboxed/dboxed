package run_infra_sandbox

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/dboxed/dboxed/pkg/sandbox"
	"github.com/dboxed/dboxed/pkg/types"
	"github.com/dboxed/dboxed/pkg/util"
)

type RunInfraSandbox struct {
	conf *types.InfraConfig

	dockerCliStdout io.WriteCloser
	dockerCliStderr io.WriteCloser
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
