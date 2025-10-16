//go:build linux

package sandbox

import (
	"context"
	"fmt"
	"time"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	run_sandbox "github.com/dboxed/dboxed/pkg/runner/run-sandbox"
	"github.com/dboxed/dboxed/pkg/runner/sandbox"
	"github.com/opencontainers/runc/libcontainer"
	"golang.org/x/sys/unix"
)

type StopCmd struct {
	SandboxName *string `help:"Specify the local sandbox name" optional:"" arg:""`

	Kill bool `help:"Send the kill signal "`
	All  bool `help:"Stop all running sandboxes"`
}

func (cmd *StopCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	sandboxBaseDir := run_sandbox.GetSandboxDir(g.WorkDir, "")

	sandboxes, err := getOneOrAllSandboxes(sandboxBaseDir, cmd.SandboxName, cmd.All)
	if err != nil {
		return err
	}

	signal := unix.SIGTERM
	if cmd.Kill {
		signal = unix.SIGKILL
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

		cs, err := s.GetSandboxContainerStatus()
		if err != nil {
			return err
		}
		if cs == libcontainer.Running && !cmd.Kill {
			err = s.StopRunInSandboxService(ctx, true)
			if err != nil {
				return err
			}
		}

		stopped, err := s.KillSandboxContainer(ctx, signal, time.Second*10)
		if err != nil {
			return err
		}
		if !stopped {
			return fmt.Errorf("failed to stop sandbox %s", si.SandboxName)
		}
		err = s.PrepareNetworkingConfig()
		if err != nil {
			return err
		}
		err = s.DestroyNetworking(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
