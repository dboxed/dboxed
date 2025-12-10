package box_spec_runner

import (
	"context"
	"log/slog"
	"strings"

	"github.com/dboxed/dboxed/pkg/runner/compose"
	"golang.org/x/sync/errgroup"
)

func (rn *BoxSpecRunner) runBoxSpecComposeUp(ctx context.Context) error {
	composeProjects, _, err := rn.loadBoxSpecComposeProjects(ctx)
	if err != nil {
		return err
	}

	var pullWg errgroup.Group
	pullWg.SetLimit(2)
	for _, p := range composeProjects {
		pullWg.Go(func() error {
			return p.RunPull(ctx)
		})
	}
	err = pullWg.Wait()
	if err != nil {
		return err
	}

	var buildWg errgroup.Group
	buildWg.SetLimit(2)
	for _, p := range composeProjects {
		buildWg.Go(func() error {
			return p.RunBuild(ctx)
		})
	}
	err = buildWg.Wait()
	if err != nil {
		return err
	}

	var upWg errgroup.Group
	upWg.SetLimit(2)
	for _, p := range composeProjects {
		upWg.Go(func() error {
			return p.RunUp(ctx, false)
		})
	}
	err = upWg.Wait()
	if err != nil {
		return err
	}
	return nil
}

func (rn *BoxSpecRunner) downDeletedBoxSpecComposeProjects(ctx context.Context) error {
	runningComposeProjects, err := compose.ListRunningComposeProjects(ctx)
	if err != nil {
		return err
	}

	composeProjects, _, err := rn.loadBoxSpecComposeProjects(ctx)
	if err != nil {
		return err
	}

	var removedComposeProjectsNames []string
	for _, cp := range runningComposeProjects {
		if strings.HasPrefix(cp.Name, "dboxed-") {
			continue
		}

		if _, ok := composeProjects[cp.Name]; ok {
			continue
		}
		removedComposeProjectsNames = append(removedComposeProjectsNames, cp.Name)
	}
	if len(removedComposeProjectsNames) != 0 {
		slog.InfoContext(ctx, "downing removed compose projects", slog.Any("composeProjects", removedComposeProjectsNames))
		for _, name := range removedComposeProjectsNames {
			err = compose.RunComposeDown(ctx, name, true, false)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (rn *BoxSpecRunner) loadBoxSpecComposeProjects(ctx context.Context) (map[string]*compose.ComposeHelper, map[string]*compose.ComposeHelper, error) {
	composeProjects, err := rn.BoxSpec.LoadComposeProjects(ctx, rn.updateServiceVolume)
	if err != nil {
		return nil, nil, err
	}
	composeProjectsOrig, err := rn.BoxSpec.LoadComposeProjects(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	ret1 := map[string]*compose.ComposeHelper{}
	ret2 := map[string]*compose.ComposeHelper{}

	for name, p := range composeProjects {
		ret1[name] = &compose.ComposeHelper{
			BaseDir:      rn.composeBaseDir,
			NameOverride: &name,
			Project:      p,
		}
	}
	for name, p := range composeProjectsOrig {
		ret2[name] = &compose.ComposeHelper{
			BaseDir:      rn.composeBaseDir,
			NameOverride: &name,
			Project:      p,
		}
	}

	return ret1, ret2, nil
}
