package box_spec_runner

import (
	"context"
	"io"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/dockercli"
	"github.com/dboxed/dboxed/pkg/types"
)

type BoxSpecRunner struct {
	DboxedVolumeLog io.Writer
}

func (rn *BoxSpecRunner) Reconcile(ctx context.Context, boxSpec *types.BoxSpec) error {
	err := rn.writeComposeFiles(ctx, boxSpec)
	if err != nil {
		return err
	}

	err = rn.reconcileVolumes(ctx, boxSpec)
	if err != nil {
		return err
	}

	err = rn.runComposeUp(ctx, boxSpec)
	if err != nil {
		return err
	}

	return nil
}

func (rn *BoxSpecRunner) Down(ctx context.Context, boxSpec *types.BoxSpec) error {
	composeProjects, err := rn.loadComposeProjects(boxSpec)
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

	var containers []types.DockerPS
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
