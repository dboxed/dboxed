//go:build linux

package sandbox

import (
	"context"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/logs"
	"github.com/dboxed/dboxed/pkg/runner/run-sandbox"
)

type RunCmd struct {
	flags.SandboxRunArgs

	WaitBeforeExit *time.Duration `help:"Wait before finally exiting. This gives the process time to print stdout/stderr messages that might be lost. Especially useful in combination with --debug"`
}

func (cmd *RunCmd) Run(g *flags.GlobalFlags, logHandler *logs.MultiLogHandler) error {
	ctx := context.Background()

	defer func() {
		if cmd.WaitBeforeExit != nil {
			slog.Info("sleeping before exit")
			time.Sleep(*cmd.WaitBeforeExit)
		}
	}()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	box, err := commandutils.GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	sandboxName, err := cmd.GetSandboxName(box)
	if err != nil {
		return err
	}

	runBox := run_sandbox.RunSandbox{
		Debug:           g.Debug,
		Client:          c,
		BoxId:           box.ID,
		InfraImage:      cmd.InfraImage,
		SandboxName:     sandboxName,
		WorkDir:         g.WorkDir,
		VethNetworkCidr: cmd.VethCidr,
	}

	err = runBox.Run(ctx, logHandler)
	if err != nil {
		return err
	}

	return nil
}
