//go:build linux

package sandbox

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/logs"
	"github.com/dboxed/dboxed/pkg/runner/run-sandbox"
)

type RunCmd struct {
	flags.SandboxRunArgs
}

func (cmd *RunCmd) Run(g *flags.GlobalFlags, logHandler *logs.MultiLogHandler) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	box, err := commandutils.GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	sandboxId, err := run_sandbox.DetermineSandboxId(ctx, c, box, g.WorkDir)
	if err != nil {
		return err
	}

	logFile := filepath.Join(run_sandbox.GetSandboxDir(g.WorkDir, sandboxId), "logs", "sandbox-run.log")
	logWriter := logs.BuildRotatingLogger(logFile)

	logHandler.AddWriter(logWriter)
	defer logHandler.RemoveWriter(logWriter)

	if !box.Enabled {
		return fmt.Errorf("the box is disabled, refusing to start it")
	}

	runBox := run_sandbox.RunSandbox{
		Debug:           g.Debug,
		Client:          c,
		BoxId:           box.ID,
		SandboxId:       sandboxId,
		InfraImage:      cmd.InfraImage,
		WorkDir:         g.WorkDir,
		VethNetworkCidr: cmd.VethCidr,
	}

	err = runBox.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}
