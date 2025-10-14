//go:build linux

package sandbox

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	run_sandbox "github.com/dboxed/dboxed/pkg/runner/run-sandbox"
	"github.com/dboxed/dboxed/pkg/runner/sandbox"
	"github.com/opencontainers/runc/libcontainer"
)

type RemoveCmd struct {
	SandboxName *string `help:"Specify the local sandbox name" optional:"" arg:""`

	All   bool `help:"Remove all sandboxes"`
	Force bool `help:"Force removal of running sandboxes. This will kill them first."`
}

func (cmd *RemoveCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	sandboxBaseDir := run_sandbox.GetSandboxDir(g.WorkDir, "")

	sandboxes, err := getOneOrAllSandboxes(sandboxBaseDir, cmd.SandboxName, cmd.All)
	if err != nil {
		return err
	}

	for _, si := range sandboxes {
		sandboxDir := run_sandbox.GetSandboxDir(g.WorkDir, si.SandboxName)

		s := sandbox.Sandbox{
			Debug:           g.Debug,
			HostWorkDir:     g.WorkDir,
			SandboxName:     si.SandboxName,
			SandboxDir:      sandboxDir,
			VethNetworkCidr: si.VethNetworkCidr,
		}

		if cmd.Force {
			err = s.StopSandboxContainer(ctx, time.Second*10)
			if err != nil {
				return err
			}
		} else {
			cs, err := s.GetSandboxContainerStatus()
			if err != nil {
				return err
			}
			if cs == libcontainer.Running {
				return fmt.Errorf("sandbox is running, stop it first (or use --force)")
			}
		}

		err = s.PrepareNetworkingConfig()
		if err != nil {
			return err
		}
		err = s.DestroyNetworking(ctx)
		if err != nil {
			slog.WarnContext(ctx,
				"destroying networking failed, but you might be able to ignore this failure",
				slog.Any("error", err.Error()),
			)
		}
		err = s.Destroy(ctx)
		if err != nil {
			return err
		}
		err = os.RemoveAll(sandboxDir)
		if err != nil {
			return err
		}
	}
	return nil
}
