package box_spec_runner

import (
	"context"
	"log/slog"
	"os"

	ctypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/runner/dockercli"
	"github.com/dboxed/dboxed/pkg/runner/network"
	"github.com/dboxed/dboxed/pkg/util"
)

type BoxSpecRunner struct {
	WorkDir      string
	BoxSpec      *boxspec.BoxSpec
	PortForwards *network.PortForwards
	Log          *slog.Logger

	NetworkIp4 *string
}

func (rn *BoxSpecRunner) Reconcile(ctx context.Context) error {
	composeProjects, composeProjectsOrig, composeBaseDir, err := rn.loadAndWriteComposeProjects(ctx)
	if err != nil {
		return err
	}
	defer os.RemoveAll(composeBaseDir)

	err = rn.downDeletedComposeProjects(ctx, composeProjects)
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

	err = rn.reconcileContentVolumes(composeProjectsOrig)
	if err != nil {
		return err
	}

	err = rn.reconcileDboxedVolumes(ctx, composeProjects, rn.BoxSpec.Volumes, true)
	if err != nil {
		return err
	}

	err = rn.runComposeUp(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (rn *BoxSpecRunner) downDeletedComposeProjects(ctx context.Context, composeProjects map[string]*ctypes.Project) error {
	runningComposeProjects, err := rn.listRunningComposeProjects(ctx)
	if err != nil {
		return err
	}

	var removedComposeProjectsNames []string
	for _, cp := range runningComposeProjects {
		if _, ok := composeProjects[cp.Name]; ok {
			continue
		}
		removedComposeProjectsNames = append(removedComposeProjectsNames, cp.Name)
	}
	if len(removedComposeProjectsNames) != 0 {
		rn.Log.InfoContext(ctx, "downing removed compose projects", slog.Any("composeProjects", removedComposeProjectsNames))
		err = rn.runComposeDownByNames(ctx, removedComposeProjectsNames, true, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (rn *BoxSpecRunner) Down(ctx context.Context, removeVolumes bool, ignoreComposeErrors bool) error {
	composeProjects, _, composeBaseDir, err := rn.loadAndWriteComposeProjects(ctx)
	if err != nil {
		return err
	}
	defer os.RemoveAll(composeBaseDir)

	err = rn.runComposeDown(ctx, composeProjects, removeVolumes, ignoreComposeErrors)
	if err != nil {
		return err
	}

	c := util.CommandHelper{
		Command: "docker",
		Args:    []string{"ps", "-a", "--format=json"},
		Logger:  rn.Log,
	}
	var containers []dockercli.DockerPS
	err = c.RunStdoutJsonLines(ctx, &containers)
	if err != nil {
		return err
	}

	var stopIds []string
	var rmIds []string
	for _, c := range containers {
		if c.State == "running" {
			stopIds = append(stopIds, c.ID)
		}
		rmIds = append(rmIds, c.ID)
	}

	if len(stopIds) != 0 {
		rn.Log.InfoContext(ctx, "stopping containers", slog.Any("ids", stopIds))
		args := []string{
			"stop",
			"--timeout=10",
		}
		args = append(args, stopIds...)
		c := util.CommandHelper{
			Command: "docker",
			Args:    args,
			Logger:  rn.Log,
			LogCmd:  true,
		}
		err = c.Run(ctx)
		if err != nil {
			return err
		}
	}
	if len(rmIds) != 0 {
		rn.Log.InfoContext(ctx, "removing containers", slog.Any("ids", stopIds))
		args := []string{
			"rm",
			"-fv",
		}
		args = append(args, rmIds...)
		c := util.CommandHelper{
			Command: "docker",
			Args:    args,
			Logger:  rn.Log,
			LogCmd:  true,
		}
		err = c.Run(ctx)
		if err != nil {
			return err
		}
	}

	rn.Log.InfoContext(ctx, "releasing dboxed volumes")
	err = rn.reconcileDboxedVolumes(ctx, composeProjects, nil, false)
	if err != nil {
		return err
	}

	return nil
}
