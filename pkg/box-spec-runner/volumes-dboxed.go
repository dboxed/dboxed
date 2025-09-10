package box_spec_runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/dboxed/dboxed/pkg/types"
)

func (rn *BoxSpecRunner) reconcileDockerVolumeDboxed(ctx context.Context, vol types.BoxVolumeSpec) error {
	workDir := rn.getVolumeWorkDir(vol)
	imageFile := filepath.Join(workDir, "image")
	snapshotMountDir := filepath.Join(workDir, "snapshot")
	lockIdFile := filepath.Join(workDir, "lock-id")

	err := os.MkdirAll(workDir, 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(snapshotMountDir, 0700)
	if err != nil {
		return err
	}

	volumeDir, err := rn.createDockerVolume(ctx, vol)
	if err != nil {
		return err
	}

	err = rn.fixVolumePermissions(vol, volumeDir)
	if err != nil {
		return err
	}

	args := []string{
		"volume",
		"serve",
		"--api-url", vol.Dboxed.ApiUrl,
		"--repo", fmt.Sprintf("%d", vol.Dboxed.RepositoryId),
		"--volume", fmt.Sprintf("%d", vol.Dboxed.VolumeId),
		"--lock-id-file", lockIdFile,
		"--image", imageFile,
		"--mount", volumeDir,
		"--snapshot-mount", snapshotMountDir,
		"--backup-interval", vol.Dboxed.BackupInterval,
	}
	env := []string{
		fmt.Sprintf("DBOXED_VOLUME_API_TOKEN=%s", vol.Dboxed.Token),
	}

	slog.Info("dboxed-volume command", slog.Any("args", strings.Join(args, " ")), slog.Any("env", strings.Join(env, ", ")))

	return nil
}
