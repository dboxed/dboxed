package box_spec_runner

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/sandbox"
	"github.com/dboxed/dboxed/pkg/types"
)

type BoxSpecRunner struct {
	Sandbox *sandbox.Sandbox
	BoxSpec *types.BoxSpec

	volumeSpecHashes []string
}

func (rn *BoxSpecRunner) Reconcile(ctx context.Context) error {
	err := rn.reconcileDockerVolumes(ctx)
	if err != nil {
		return err
	}

	err = rn.runComposeUp(ctx)
	if err != nil {
		return err
	}

	return nil
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

	var containers []types.DockerPS
	err = rn.Sandbox.RunDockerCliJsonLines(ctx, &containers, "", "ps", "-a", "--format=json")
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
		_, err = rn.Sandbox.RunDockerCli(ctx, false, "", args...)
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
		_, err = rn.Sandbox.RunDockerCli(ctx, false, "", args...)
		if err != nil {
			return err
		}
	}

	return nil
}
