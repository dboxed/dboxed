package box_spec_runner

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	ctypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/dboxed/dboxed/pkg/runner/dockercli"
	"github.com/dboxed/dboxed/pkg/util"
	"golang.org/x/sync/errgroup"
)

func (rn *BoxSpecRunner) writeComposeFiles(dir string, composeProjects map[string]*ctypes.Project) error {
	for name, composeProject := range composeProjects {
		b, err := composeProject.MarshalYAML()
		if err != nil {
			return err
		}

		dir := filepath.Join(dir, name)
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return err
		}
		err = os.WriteFile(filepath.Join(dir, "docker-compose.yaml"), b, 0600)
		if err != nil {
			return err
		}
	}
	return nil
}

func (rn *BoxSpecRunner) listRunningComposeProjects(ctx context.Context) ([]dockercli.DockerComposeListEntry, error) {
	cmd := util.CommandHelper{
		Command: "docker",
		Args:    []string{"compose", "ls", "-a", "--format", "json"},
		Logger:  rn.Log,
	}
	var l []dockercli.DockerComposeListEntry
	err := cmd.RunStdoutJson(ctx, &l)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (rn *BoxSpecRunner) runComposeUp(ctx context.Context) error {
	composeProjects, composeBaseDir, err := rn.loadAndWriteComposeProjects(ctx)
	if err != nil {
		return err
	}
	defer os.RemoveAll(composeBaseDir)

	var pullWg errgroup.Group
	pullWg.SetLimit(2)
	for name := range composeProjects {
		pullWg.Go(func() error {
			return rn.runComposeCli(ctx, composeBaseDir, name, nil, "pull")
		})
	}
	err = pullWg.Wait()
	if err != nil {
		return err
	}

	var buildWg errgroup.Group
	buildWg.SetLimit(2)
	for name := range composeProjects {
		buildWg.Go(func() error {
			return rn.runComposeCli(ctx, composeBaseDir, name, nil, "build")
		})
	}
	err = buildWg.Wait()
	if err != nil {
		return err
	}

	var upWg errgroup.Group
	upWg.SetLimit(2)
	for name := range composeProjects {
		upWg.Go(func() error {
			return rn.runComposeCli(ctx, composeBaseDir, name, nil, "up", "-d", "--remove-orphans", "--pull=never")
		})
	}
	err = upWg.Wait()
	if err != nil {
		return err
	}
	return nil
}

func (rn *BoxSpecRunner) runComposeDown(ctx context.Context, composeProjects map[string]*ctypes.Project, removeVolumes bool, ignoreComposeErrors bool) error {
	var names []string
	for name := range composeProjects {
		names = append(names, name)
	}
	return rn.runComposeDownByNames(ctx, names, removeVolumes, ignoreComposeErrors)
}

func (rn *BoxSpecRunner) runComposeDownByNames(ctx context.Context, names []string, removeVolumes bool, ignoreComposeErrors bool) error {
	var wg errgroup.Group
	wg.SetLimit(2)
	for _, name := range names {
		wg.Go(func() error {
			args := []string{
				"-p", name,
				"down", "--remove-orphans",
			}
			if removeVolumes {
				args = append(args, "-v")
			}
			err := rn.runComposeCli(ctx, "", name, nil, args...)
			if err != nil {
				rn.Log.ErrorContext(ctx, "error while calling docker compose", slog.Any("error", err))
				if !ignoreComposeErrors {
					return err
				}
			}
			return nil
		})
	}
	err := wg.Wait()
	if err != nil {
		return err
	}
	return nil
}

func (rn *BoxSpecRunner) loadAndWriteComposeProjects(ctx context.Context) (map[string]*ctypes.Project, string, error) {
	getMount := func(volumeId string) string {
		return rn.getVolumeMountDir(volumeId)
	}
	composeProjects, err := rn.BoxSpec.LoadComposeProjects(ctx, getMount)
	if err != nil {
		return nil, "", err
	}

	tmpDir, err := os.MkdirTemp("", "dboxed-docker-compose-")
	if err != nil {
		return nil, "", err
	}

	err = rn.writeComposeFiles(tmpDir, composeProjects)
	if err != nil {
		return nil, "", err
	}
	return composeProjects, tmpDir, nil
}

func (rn *BoxSpecRunner) runComposeCli(ctx context.Context, composeBaseDir string, projectName string, cmdEnv []string, args ...string) error {
	var args2 []string
	args2 = append(args2, "compose")
	args2 = append(args2, args...)

	dir := ""
	if composeBaseDir != "" {
		dir = filepath.Join(composeBaseDir, projectName)
	}

	cmd := util.CommandHelper{
		Command: "docker",
		Args:    args2,
		Env:     cmdEnv,
		Dir:     dir,
		Logger:  rn.Log.With(slog.Any("composeProject", projectName)),
		LogCmd:  true,
	}
	err := cmd.Run(ctx)
	if err != nil {
		return err
	}
	return nil
}
