package box_spec_runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	ctypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/dboxed/dboxed/pkg/types"
)

func (rn *BoxSpecRunner) runComposeUp(ctx context.Context) error {
	projectPathInSandbox := filepath.Join(types.DboxedDataDir, "compose")
	projectPathOnHost := filepath.Join(rn.Sandbox.GetSandboxRoot(), projectPathInSandbox)

	err := os.MkdirAll(projectPathOnHost, 0700)
	if err != nil {
		return err
	}

	composeProjects, err := rn.loadComposeProjects()
	if err != nil {
		return err
	}

	for _, composeProject := range composeProjects {
		b, err := composeProject.MarshalYAML()
		if err != nil {
			return err
		}

		dir := rn.buildComposeDir(true, composeProject.Name)
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return err
		}
		err = os.WriteFile(filepath.Join(dir, "docker-compose.yaml"), b, 0600)
		if err != nil {
			return err
		}
	}

	for _, composeProject := range composeProjects {
		err = rn.runComposeCli(ctx, composeProject.Name, "pull", "-q")
		if err != nil {
			return err
		}
	}

	for _, composeProject := range composeProjects {
		err = rn.runComposeCli(ctx, composeProject.Name, "up", "-d", "--remove-orphans")
		if err != nil {
			return err
		}
	}
	return nil
}

func (rn *BoxSpecRunner) loadComposeProjects() ([]*ctypes.Project, error) {
	composeProjects, err := rn.BoxSpec.LoadComposeProjects()
	if err != nil {
		return nil, err
	}
	for i, composeProject := range composeProjects {
		if composeProject.Name == "" {
			composeProject.Name = fmt.Sprintf("tmp-%d", i)
		}
		err = rn.setupComposeFile(composeProject)
		if err != nil {
			return nil, err
		}
	}

	return composeProjects, nil
}

func (rn *BoxSpecRunner) buildComposeDir(host bool, name string) string {
	projectPathInSandbox := filepath.Join(types.DboxedDataDir, "compose")
	dir := filepath.Join(projectPathInSandbox, name)
	if host {
		dir = filepath.Join(rn.Sandbox.GetSandboxRoot(), dir)
	}
	return dir
}

func (rn *BoxSpecRunner) runComposeCli(ctx context.Context, projectName string, args ...string) error {
	log := slog.With("composeProject", projectName)

	var args2 []string
	args2 = append(args2, "compose")
	args2 = append(args2, args...)

	slog.InfoContext(ctx, fmt.Sprintf("running 'docker compose %s'", strings.Join(args, " ")), slog.Any("projectName", projectName))
	_, err := rn.Sandbox.RunDockerCli(ctx, log, false, rn.buildComposeDir(false, projectName), args2...)
	if err != nil {
		return err
	}
	return nil
}

func (rn *BoxSpecRunner) setupComposeFile(compose *ctypes.Project) error {
	if compose.Volumes == nil {
		compose.Volumes = map[string]ctypes.VolumeConfig{}
	}

	volumesByName := map[string]int{}
	for i, vol := range rn.BoxSpec.Volumes {
		volumesByName[vol.Name] = i

		volumeName := rn.getDockerVolumeName(vol)
		compose.Volumes[volumeName] = ctypes.VolumeConfig{
			Name:     volumeName,
			External: true,
		}
	}

	for _, service := range compose.Services {
		for i, _ := range service.Volumes {
			volume := &service.Volumes[i]
			if volume.Type == "dboxed" {
				volIdx, ok := volumesByName[volume.Source]
				if !ok {
					return fmt.Errorf("volume with name %s not found", volume.Source)
				}
				vol := rn.BoxSpec.Volumes[volIdx]
				volume.Type = "volume"
				volume.Source = rn.getDockerVolumeName(vol)
				volume.ReadOnly = true
			}
		}
	}
	return nil
}
