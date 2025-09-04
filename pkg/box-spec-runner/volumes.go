package box_spec_runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/dboxed/dboxed-common/util"
	"github.com/dboxed/dboxed/pkg/types"
)

func (rn *BoxSpecRunner) getVolumeWorkDirOnHost(vol types.BoxVolumeSpec) string {
	return filepath.Join(rn.Sandbox.SandboxDir, "volumes", vol.Uuid)
}

func (rn *BoxSpecRunner) getVolumeMountDirOnHost(vol types.BoxVolumeSpec) string {
	return filepath.Join(rn.getVolumeWorkDirOnHost(vol), "mount")
}

func (rn *BoxSpecRunner) getVolumeWorkDirInSandbox(vol types.BoxVolumeSpec) string {
	return filepath.Join(types.VolumesDir, vol.Uuid)
}

func (rn *BoxSpecRunner) getVolumeMountDirInSandbox(vol types.BoxVolumeSpec) string {
	return filepath.Join(rn.getVolumeWorkDirInSandbox(vol), "mount")
}

func (rn *BoxSpecRunner) getDockerVolumeName(vol types.BoxVolumeSpec) string {
	h := util.MustSha256SumJson(vol)
	return fmt.Sprintf("dboxed-%s-%s", vol.Name, h[:6])
}

func (rn *BoxSpecRunner) createDockerVolume(ctx context.Context, vol types.BoxVolumeSpec) error {
	volumeName := rn.getDockerVolumeName(vol)
	mountDir := rn.getVolumeMountDirInSandbox(vol)

	slog.InfoContext(ctx, "creating docker volume", slog.Any("volumeName", volumeName))
	_, err := rn.Sandbox.RunDockerCli(ctx, slog.Default(), true, "", "volume", "create", volumeName)
	if err != nil {
		return err
	}

	volumeDir, err := rn.getDockerVolumeDirInSandbox(ctx, vol)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "bind mounting volume",
		slog.Any("volumeName", volumeName),
		slog.Any("mountDir", mountDir),
		slog.Any("volumeDir", volumeDir),
	)
	_, err = rn.Sandbox.RunSandboxCmd(ctx, slog.Default(), false, "", "mount", "-obind", mountDir, volumeDir)
	if err != nil {
		return err
	}

	return nil
}

func (rn *BoxSpecRunner) getDockerVolumeDirInSandbox(ctx context.Context, vol types.BoxVolumeSpec) (string, error) {
	volumeName := rn.getDockerVolumeName(vol)

	var volumeInfos []types.DockerVolume
	err := rn.Sandbox.RunDockerCliJson(ctx, slog.Default(), &volumeInfos, "", "volume", "inspect", volumeName, "--format", "json")
	if err != nil {
		return "", err
	}

	path := volumeInfos[0].Mountpoint
	return path, nil
}

func (rn *BoxSpecRunner) reconcileDockerVolumes(ctx context.Context) error {
	for _, vol := range rn.BoxSpec.Volumes {
		workDir := rn.getVolumeWorkDirOnHost(vol)
		mountDir := rn.getVolumeMountDirOnHost(vol)

		err := os.MkdirAll(workDir, 0700)
		if err != nil {
			return err
		}

		err = os.MkdirAll(mountDir, 0700)
		if err != nil {
			return err
		}

		if vol.FileBundle != nil {
			err = rn.createFileBundle(ctx, vol)
			if err != nil {
				return err
			}
		} else if vol.Dboxed != nil {
			err = rn.createDboxedVolume(ctx, vol)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("missing volume config")
		}

		err = rn.createDockerVolume(ctx, vol)
		if err != nil {
			return err
		}

		err = rn.fixVolumePermissions(vol)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rn *BoxSpecRunner) fixVolumePermissions(vol types.BoxVolumeSpec) error {
	mountDir := rn.getVolumeMountDirOnHost(vol)
	err := os.Chown(mountDir, int(vol.RootUid), int(vol.RootGid))
	if err != nil {
		return err
	}
	rootMode, err := parseMode(vol.RootMode)
	if err != nil {
		return fmt.Errorf("failed to parse root dir mode: %w", err)
	}
	if rootMode != 0 {
		err = os.Chmod(mountDir, rootMode)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseMode(s string) (os.FileMode, error) {
	n, err := strconv.ParseInt(s, 8, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid file mode %s: %w", s, err)
	}
	fm := os.FileMode(n)
	if fm & ^os.ModePerm != 0 {
		return 0, fmt.Errorf("invalid file mode %s", s)
	}
	return fm, nil
}
