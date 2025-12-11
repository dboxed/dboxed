package box_spec_runner

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/runner/compose"
	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/runner/dockercli"
	"github.com/dboxed/dboxed/pkg/runner/network"
	"github.com/dboxed/dboxed/pkg/util"
)

type BoxSpecRunner struct {
	Client       *baseclient.Client
	BoxSpec      *boxspec.BoxSpec
	PortForwards *network.PortForwards

	NetworkIp4 *string

	composeBaseDir string
}

func (rn *BoxSpecRunner) Reconcile(ctx context.Context) error {
	composeBaseDir := filepath.Join(consts.DboxedDataDir, "compose")
	err := os.MkdirAll(composeBaseDir, 0700)
	if err != nil {
		return err
	}
	rn.composeBaseDir = composeBaseDir

	err = rn.downDeletedBoxSpecComposeProjects(ctx)
	if err != nil {
		return err
	}

	err = rn.reconcilePortForwards(ctx)
	if err != nil {
		return err
	}
	err = rn.reconcileNetwork(ctx)
	if err != nil {
		return err
	}

	err = rn.reconcileContentVolumes(ctx)
	if err != nil {
		return err
	}

	err = rn.reconcileDboxedVolumes(ctx, true)
	if err != nil {
		return err
	}

	err = rn.runBoxSpecComposeUp(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (rn *BoxSpecRunner) Down(ctx context.Context, removeVolumes bool, ignoreComposeErrors bool) error {
	composeProjects, _, err := rn.loadBoxSpecComposeProjects(ctx)
	if err != nil {
		return err
	}

	for name, _ := range composeProjects {
		err = compose.RunComposeDown(ctx, name, removeVolumes, ignoreComposeErrors)
		if err != nil {
			return err
		}
	}

	err = rn.downDboxedVolumes(ctx)
	if err != nil {
		return err
	}

	runningComposeProjects, err := compose.ListRunningComposeProjects(ctx)
	if err != nil {
		return err
	}
	for _, p := range runningComposeProjects {
		err = compose.RunComposeDown(ctx, p.Name, removeVolumes, ignoreComposeErrors)
		if err != nil {
			return err
		}
	}

	err = rn.downAllContainers(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (rn *BoxSpecRunner) downAllContainers(ctx context.Context) error {
	c := util.CommandHelper{
		Command: "docker",
		Args:    []string{"ps", "-a", "--format=json"},
		Logger:  slog.Default(),
	}
	var containers []dockercli.DockerPS
	err := c.RunStdoutJsonLines(ctx, &containers)
	if err != nil {
		return err
	}

	var stopIds []string
	var rmIds []string
	for _, c := range containers {
		if c.Names == "dboxed-run-in-sandbox" {
			continue
		}

		if c.State == "running" {
			stopIds = append(stopIds, c.ID)
		}
		rmIds = append(rmIds, c.ID)
	}

	if len(stopIds) != 0 {
		slog.InfoContext(ctx, "stopping containers", slog.Any("ids", stopIds))
		args := []string{
			"stop",
			"--timeout=10",
		}
		args = append(args, stopIds...)
		c := util.CommandHelper{
			Command: "docker",
			Args:    args,
			Logger:  slog.Default(),
			LogCmd:  true,
		}
		err = c.Run(ctx)
		if err != nil {
			return err
		}
	}
	if len(rmIds) != 0 {
		slog.InfoContext(ctx, "removing containers", slog.Any("ids", stopIds))
		args := []string{
			"rm",
			"-fv",
		}
		args = append(args, rmIds...)
		c := util.CommandHelper{
			Command: "docker",
			Args:    args,
			Logger:  slog.Default(),
			LogCmd:  true,
		}
		err = c.Run(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
