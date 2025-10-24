//go:build linux

package sandbox

import (
	"context"
	"fmt"

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

	if box.DesiredState != "up" {
		return fmt.Errorf("the desired state of the box is '%s', refusing to start it", box.DesiredState)
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
