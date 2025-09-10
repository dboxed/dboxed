package box_spec_runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	ctypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/dboxed/dboxed/pkg/types"
)

func (rn *BoxSpecRunner) reconcileDockerVolumeDboxed(ctx context.Context, vol types.BoxVolumeSpec, volumesProject *ctypes.Project) error {
	workDirOnHost := rn.getVolumeWorkDirOnHost(vol)
	workDirInSandbox := rn.getVolumeWorkDirInSandbox(vol)
	volumeName := rn.getDockerVolumeName(vol)

	err := os.MkdirAll(workDirOnHost, 0700)
	if err != nil {
		return err
	}

	volumeDirInSandbox, volumeDirOnHost, err := rn.createDockerVolume(ctx, vol)
	if err != nil {
		return err
	}

	err = rn.fixVolumePermissions(vol, volumeDirOnHost)
	if err != nil {
		return err
	}

	cmd := []string{
		"volume",
		"serve",
		"--api-url", vol.Dboxed.ApiUrl,
		"--repo", fmt.Sprintf("%d", vol.Dboxed.RepositoryId),
		"--volume", fmt.Sprintf("%d", vol.Dboxed.VolumeId),
		"--lock-id-file", "/volume/lock-id",
		"--image", "/volume/image",
		"--mount", "/volume/mount",
		"--snapshot-mount", "/volume/snapshot",
		"--backup-interval", vol.Dboxed.BackupInterval,
	}
	slog.Info("dboxed-volume command", slog.Any("cmd", strings.Join(cmd, " ")))

	volumesProject.Services[volumeName] = ctypes.ServiceConfig{
		Name:       volumeName,
		Image:      "ghcr.io/dboxed/dboxed-volume:nightly",
		PullPolicy: "always",
		Restart:    "on-failure",
		Privileged: true,
		Entrypoint: []string{"sleep", "infinity"},
		//Command:    cmd,
		Environment: map[string]*string{
			"DBOXED_VOLUME_API_TOKEN": &vol.Dboxed.Token,
		},
		Volumes: []ctypes.ServiceVolumeConfig{
			{
				Type:   "bind",
				Source: "/dev",
				Target: "/dev",
			},
			{
				Type:   "bind",
				Source: workDirInSandbox,
				Target: "/volume",
			},
			{
				Type:   "bind",
				Source: volumeDirInSandbox,
				Target: "/volume/mount",
			},
		},
	}

	return nil
}
