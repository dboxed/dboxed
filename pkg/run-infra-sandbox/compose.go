package run_infra_sandbox

import (
	"context"
	"fmt"
	ctypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/koobox/unboxed/pkg/types"
	"os"
	"path/filepath"
)

func (rn *RunInfraSandbox) runComposeUp(ctx context.Context) error {
	projectPath := filepath.Join(types.UnboxedDataDir, "compose")
	configPath := filepath.Join(projectPath, "compose.yaml")

	err := os.MkdirAll(projectPath, 0700)
	if err != nil {
		return err
	}

	compose, err := rn.conf.BoxSpec.LoadComposeProject()
	if err != nil {
		return err
	}

	err = rn.setupComposeFile(ctx, compose)
	if err != nil {
		return err
	}

	b, err := compose.MarshalYAML()
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, b, 0600)
	if err != nil {
		return err
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

	for k, service := range compose.Services {
		if service.NetworkMode == "" && len(service.Networks) == 0 {
			service.NetworkMode = fmt.Sprintf("ns:/run/netns/%s", rn.network.NamesAndIps.SandboxNamespaceName)
		}
		if len(service.DNS) == 0 {
			service.DNS = append(service.DNS, rn.conf.NetworkConfig.DnsProxyIP)
		}
		compose.Services[k] = service
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
