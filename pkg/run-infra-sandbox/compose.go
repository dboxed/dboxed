package run_infra_sandbox

import (
	"context"
	"fmt"
	ctypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/dboxed/dboxed/pkg/types"
	"log/slog"
	"os"
	"path/filepath"
)

func (rn *RunInfraSandbox) runComposeUp(ctx context.Context) error {
	projectPath := filepath.Join(types.DboxedDataDir, "compose")

	err := os.MkdirAll(projectPath, 0700)
	if err != nil {
		return err
	}

	composeProjects, err := rn.conf.BoxSpec.LoadComposeProjects()
	if err != nil {
		return err
	}

	for i, composeProject := range composeProjects {
		err = rn.setupComposeFile(ctx, composeProject)
		if err != nil {
			return err
		}

		if composeProject.Name == "" {
			composeProject.Name = fmt.Sprintf("tmp-%d", i)
		}

		b, err := composeProject.MarshalYAML()
		if err != nil {
			return err
		}

		configName := fmt.Sprintf("compose-%s.yaml", composeProject.Name)
		err = os.WriteFile(filepath.Join(projectPath, configName), b, 0600)
		if err != nil {
			return err
		}
	}

	for _, composeProject := range composeProjects {
		configFile := fmt.Sprintf("compose-%s.yaml", composeProject.Name)
		cmd := rn.buildDockerCliCmd(ctx, "compose", "-f", configFile, "pull", "-q")
		cmd.Dir = projectPath
		err = cmd.Run()
		if err != nil {
			return err
		}
	}

	for _, composeProject := range composeProjects {
		slog.InfoContext(ctx, "running docker compose up", slog.Any("projectName", composeProject.Name))
		configFile := fmt.Sprintf("compose-%s.yaml", composeProject.Name)
		cmd := rn.buildDockerCliCmd(ctx, "compose", "-f", configFile, "up", "-d")
		cmd.Dir = projectPath
		err = cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func (rn *RunInfraSandbox) setupComposeFile(ctx context.Context, compose *ctypes.Project) error {
	if compose.Volumes == nil {
		compose.Volumes = map[string]ctypes.VolumeConfig{}
	}

	bundlesByName := map[string]*types.FileBundle{}
	for _, fb := range rn.conf.BoxSpec.FileBundles {
		bundlesByName[fb.Name] = &fb

		volumeName := rn.getBundleVolumeName(fb.Name)
		compose.Volumes[volumeName] = ctypes.VolumeConfig{
			Name:     volumeName,
			External: true,
		}
	}

	for _, service := range compose.Services {
		for i, _ := range service.Volumes {
			volume := &service.Volumes[i]
			if volume.Type == "bundle" {
				fb := bundlesByName[volume.Source]
				if fb == nil {
					return fmt.Errorf("file bundle with name %s not found", volume.Source)
				}
				volume.Type = "volume"
				volume.Source = rn.getBundleVolumeName(fb.Name)
			}
		}
	}
	return nil
}
