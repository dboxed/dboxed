package box_spec_runner

import (
	"context"
	"log/slog"
	"os"

	ctypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/runner/dockercli"
	"github.com/dboxed/dboxed/pkg/util"
)

type BoxSpecRunner struct {
	WorkDir string
	BoxSpec *boxspec.BoxSpec
}

func (rn *BoxSpecRunner) Reconcile(ctx context.Context) error {
	composeProjects, composeBaseDir, err := rn.loadAndWriteComposeProjects(ctx)
	if err != nil {
		return err
	}
	defer os.RemoveAll(composeBaseDir)

	err = rn.downDeletedComposeProjects(ctx, composeProjects)
	if err != nil {
		return err
	}

	err = rn.reconcileVolumes(ctx, composeProjects, rn.BoxSpec.Volumes, true)
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
		slog.InfoContext(ctx, "downing removed compose projects", slog.Any("composeProjects", removedComposeProjectsNames))
		err = rn.runComposeDownByNames(ctx, removedComposeProjectsNames, true, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (rn *BoxSpecRunner) Down(ctx context.Context, removeVolumes bool, ignoreComposeErrors bool) error {
	composeProjects, composeBaseDir, err := rn.loadAndWriteComposeProjects(ctx)
	if err != nil {
		return err
	}
	defer os.RemoveAll(composeBaseDir)

	err = rn.runComposeDown(ctx, composeProjects, removeVolumes, ignoreComposeErrors)
	if err != nil {
		return err
	}

	containers, err := util.RunCommandJsonLines[dockercli.DockerPS](ctx, "docker", "ps", "-a", "--format=json")
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
		slog.InfoContext(ctx, "stopping containers", slog.Any("ids", stopIds))
		args := []string{
			"stop",
			"--timeout=10",
		}
		args = append(args, stopIds...)
		err = util.RunCommand(ctx, "docker", args...)
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
		err = util.RunCommand(ctx, "docker", args...)
		if err != nil {
			return err
		}
	}

	slog.InfoContext(ctx, "releasing dboxed volumes")
	err = rn.reconcileVolumes(ctx, composeProjects, nil, false)
	if err != nil {
		return err
	}

	return nil
}
