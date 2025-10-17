package box_spec_runner

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/runner/dockercli"
	"github.com/dboxed/dboxed/pkg/util"
)

type BoxSpecRunner struct {
	WorkDir string
	BoxSpec *boxspec.BoxSpec
}

func (rn *BoxSpecRunner) Reconcile(ctx context.Context) error {
	err := rn.writeComposeFiles(ctx)
	if err != nil {
		return err
	}

	err = rn.reconcileVolumes(ctx, rn.BoxSpec.Volumes, true)
	if err != nil {
		return err
	}

	err = rn.runComposeUp(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (rn *BoxSpecRunner) Down(ctx context.Context, ignoreComposeErrors bool) error {
	composeProjects, err := rn.loadComposeProjects(ctx)
	if err != nil {
		return err
	}

	for i := len(composeProjects) - 1; i >= 0; i-- {
		composeProject := composeProjects[i]
		err = rn.runComposeCli(ctx, composeProject.Name, nil, "down", "-v", "--remove-orphans", "--timeout=10")
		if err != nil {
			if !ignoreComposeErrors {
				return err
			}
			slog.ErrorContext(ctx, "error while calling docker compose", slog.Any("error", err))
		}
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
	err = rn.reconcileVolumes(ctx, nil, false)
	if err != nil {
		return err
	}

	return nil
}
