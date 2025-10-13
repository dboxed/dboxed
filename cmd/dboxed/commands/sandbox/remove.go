//go:build linux

package sandbox

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/runner/sandbox"
)

type RemoveCmd struct {
	SandboxName *string `help:"Specify the local sandbox name" optional:"" arg:""`

	All   bool `help:"Remove all sandboxes"`
	Force bool `help:"Force removal of running sandboxes. This will kill them first."`
}

func (cmd *RemoveCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	sandboxBaseDir := filepath.Join(g.WorkDir, "sandboxes")

	sandboxes, err := getOneOrAllSandboxes(sandboxBaseDir, cmd.SandboxName, cmd.All)
	if err != nil {
		return err
	}

	for _, si := range sandboxes {
		sandboxDir := filepath.Join(g.WorkDir, "sandboxes", si.SandboxName)

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
