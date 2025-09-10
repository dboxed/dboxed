package box_spec_runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/dboxed/dboxed-common/util"
	"github.com/dboxed/dboxed/pkg/dockercli"
	"github.com/dboxed/dboxed/pkg/types"
)

func (rn *BoxSpecRunner) getVolumeWorkDirBase(vol types.BoxVolumeSpec) string {
	return rn.getDockerVolumeName(vol)
}

func (rn *BoxSpecRunner) getVolumeWorkDir(vol types.BoxVolumeSpec) string {
	return filepath.Join(types.VolumesDir, rn.getVolumeWorkDirBase(vol))
}

func (rn *BoxSpecRunner) getDockerVolumeName(vol types.BoxVolumeSpec) string {
	h := util.MustSha256SumJson(vol)
	if vol.FileBundle != nil {
		return fmt.Sprintf("file-bundle-%s-%s", vol.Name, h[:6])
	} else if vol.Dboxed != nil {
		return fmt.Sprintf("dboxed-volume-%d-%d", vol.Dboxed.RepositoryId, vol.Dboxed.VolumeId)
	} else {
		panic("volume type not supported")
	}
}

func (rn *BoxSpecRunner) createDockerVolume(ctx context.Context, vol types.BoxVolumeSpec) (string, error) {
	volumeName := rn.getDockerVolumeName(vol)

	slog.InfoContext(ctx, "creating docker volume", slog.Any("volumeName", volumeName))
	_, err := dockercli.RunDockerCli(ctx, slog.Default(), true, "", "volume", "create", volumeName)
	if err != nil {
		return "", err
	}

	volumeDir, err := rn.getDockerVolumeDir(ctx, vol)
	if err != nil {
		return "", err
	}

	return volumeDir, nil
}

func (rn *BoxSpecRunner) getDockerVolumeDir(ctx context.Context, vol types.BoxVolumeSpec) (string, error) {
	volumeName := rn.getDockerVolumeName(vol)

	var volumeInfos []types.DockerVolume
	err := dockercli.RunDockerCliJson(ctx, slog.Default(), &volumeInfos, "", "volume", "inspect", volumeName, "--format", "json")
	if err != nil {
		return "", err
	}

	path := volumeInfos[0].Mountpoint
	return path, nil
}

func (rn *BoxSpecRunner) reconcileDockerVolumes(ctx context.Context) error {
	for _, vol := range rn.BoxSpec.Volumes {
		if vol.FileBundle != nil {
			err := rn.reconcileDockerVolumeFileBundle(ctx, vol)
			if err != nil {
				return err
			}
		} else if vol.Dboxed != nil {
			err := rn.reconcileDockerVolumeDboxed(ctx, vol)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("volume %s has unsupported volume type", vol.Name)
		}
	}

	return nil
}

func (rn *BoxSpecRunner) fixVolumePermissions(vol types.BoxVolumeSpec, volumeDir string) error {
	err := os.Chown(volumeDir, int(vol.RootUid), int(vol.RootGid))
	if err != nil {
		return err
	}
	rootMode, err := parseMode(vol.RootMode)
	if err != nil {
		return fmt.Errorf("failed to parse root dir mode: %w", err)
	}
	if rootMode != 0 {
		err = os.Chmod(volumeDir, rootMode)
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
