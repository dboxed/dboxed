package box_spec_runner

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	ctypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/util"
)

func (rn *BoxSpecRunner) writeComposeFiles(ctx context.Context) error {
	composeProjects, err := rn.loadComposeProjects(ctx)
	if err != nil {
		return err
	}

	for _, composeProject := range composeProjects {
		b, err := composeProject.MarshalYAML()
		if err != nil {
			return err
		}

		dir := rn.buildComposeDir(composeProject.Name)
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

func (rn *BoxSpecRunner) runComposeUp(ctx context.Context) error {
	composeProjects, err := rn.loadComposeProjects(ctx)
	if err != nil {
		return err
	}

	for _, composeProject := range composeProjects {
		err = rn.runComposeCli(ctx, composeProject.Name, nil, "pull", "-q")
		if err != nil {
			return err
		}
	}

	for _, composeProject := range composeProjects {
		err = rn.runComposeCli(ctx, composeProject.Name, nil, "build", "-q")
		if err != nil {
			return err
		}
	}

	for _, composeProject := range composeProjects {
		err = rn.runComposeCli(ctx, composeProject.Name, nil, "up", "-d", "--remove-orphans", "--pull=never")
		if err != nil {
			return err
		}
	}
	return nil
}

func (rn *BoxSpecRunner) runComposeDown(ctx context.Context) error {
	composeProjects, err := rn.loadComposeProjects(ctx)
	if err != nil {
		return err
	}

	for _, composeProject := range composeProjects {
		err = rn.runComposeCli(ctx, composeProject.Name, nil, "down", "--remove-orphans")
		if err != nil {
			return err
		}
	}
	return nil
}

func (rn *BoxSpecRunner) loadComposeProjects(ctx context.Context) ([]*ctypes.Project, error) {
	getMount := func(volumeUuid string) string {
		return rn.getVolumeMountDir(volumeUuid)
	}

	return rn.BoxSpec.LoadComposeProjects(ctx, getMount)
}

func (rn *BoxSpecRunner) buildComposeDir(name string) string {
	projectPathInSandbox := filepath.Join(consts.DboxedDataDir, "compose")
	dir := filepath.Join(projectPathInSandbox, name)
	return dir
}

func (rn *BoxSpecRunner) runComposeCli(ctx context.Context, projectName string, cmdEnv []string, args ...string) error {
	var args2 []string
	args2 = append(args2, "compose")
	args2 = append(args2, args...)

	cmd := util.CommandHelper{
		Command: "docker",
		Args:    args2,
		Env:     cmdEnv,
		Dir:     rn.buildComposeDir(projectName),
		Logger:  slog.Default(),
		LogCmd:  true,
	}
	err := cmd.Run(ctx)
	if err != nil {
		return err
	}
	return nil
}
