//go:build linux

package sandbox

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/runner/network"
	run_sandbox "github.com/dboxed/dboxed/pkg/runner/run-sandbox"
	"github.com/dboxed/dboxed/pkg/runner/sandbox"
	"github.com/opencontainers/runc/libcontainer"
)

type RemoveCmd struct {
	flags.SandboxArgsOptional

	All   bool `help:"Remove all sandboxes"`
	Force bool `help:"Force removal of running sandboxes. This will kill them first."`
}

func (cmd *RemoveCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}
	bc := clients.BoxClient{Client: c}

	sandboxBaseDir := run_sandbox.GetSandboxDir(g.WorkDir, "")
	sandboxes, err := commandutils.GetOneOrAllSandboxInfos(sandboxBaseDir, cmd.Sandbox, cmd.All)
	if err != nil {
		return err
	}

	for _, si := range sandboxes {
		sandboxDir := run_sandbox.GetSandboxDir(g.WorkDir, si.SandboxId)

		box, err := bc.GetBoxById(ctx, si.Box.ID)
		if err != nil {
			return err
		}

		s := sandbox.Sandbox{
			Debug:       g.Debug,
			HostWorkDir: g.WorkDir,
			SandboxId:   si.SandboxId,
			SandboxDir:  sandboxDir,
		}

		cs, err := s.GetSandboxContainerStatus()
		if err != nil {
			return err
		}

		if cmd.Force {
			if cs == libcontainer.Running {
				err = s.StopRunInSandboxService(ctx, true)
				if err != nil {
					return err
				}
			}

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

		namesAndIps, err := network.NewNamesAndIPs(si.SandboxId, si.AcquiredVethNetworkCidr)
		if err != nil {
			return err
		}

		err = network.Destroy(ctx, nil, namesAndIps, s.GetSandboxRoot())
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

		if box.Sandbox != nil && box.Sandbox.ID == si.SandboxId {
			slog.InfoContext(ctx, "releasing box sandbox")
			err = bc.ReleaseSandbox(ctx, box.ID, si.SandboxId)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
