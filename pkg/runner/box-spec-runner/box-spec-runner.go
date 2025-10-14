package box_spec_runner

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/runner/dockercli"
)

type BoxSpecRunner struct {
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

func (rn *BoxSpecRunner) DownVolumes(ctx context.Context) error {
	return rn.reconcileVolumes(ctx, nil, false)
}

func (rn *BoxSpecRunner) Down(ctx context.Context) error {
	composeProjects, err := rn.loadComposeProjects()
	if err != nil {
		return err
	}

	for i := len(composeProjects) - 1; i >= 0; i-- {
		composeProject := composeProjects[i]
		err = rn.runComposeCli(ctx, composeProject.Name, "down", "-v", "--remove-orphans", "--timeout=10")
		if err != nil {
			return err
		}
	}

	var containers []dockercli.DockerPS
	err = dockercli.RunDockerCliJsonLines(ctx, slog.Default(), &containers, "", "ps", "-a", "--format=json")
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
		_, err = dockercli.RunDockerCli(ctx, slog.Default(), false, "", args...)
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
		_, err = dockercli.RunDockerCli(ctx, slog.Default(), false, "", args...)
		if err != nil {
			return err
		}
	}

	return nil
}
