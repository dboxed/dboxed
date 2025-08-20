package box_spec_runner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/dboxed/dboxed/pkg/types"
)

func (rn *BoxSpecRunner) getDockerVolumeName(name string, specHash string) string {
	return fmt.Sprintf("dboxed-%s-%s", name, specHash[:6])
}

func (rn *BoxSpecRunner) createDockerVolume(ctx context.Context, name string, specHash string) (string, error) {
	volumeName := rn.getDockerVolumeName(name, specHash)

	slog.InfoContext(ctx, "creating docker volume", slog.Any("volumeName", volumeName))
	_, err := rn.Sandbox.RunDockerCli(ctx, slog.Default(), true, "", "volume", "create", volumeName)
	if err != nil {
		return "", err
	}

	var volumeInfos []types.DockerVolume
	err = rn.Sandbox.RunDockerCliJson(ctx, slog.Default(), &volumeInfos, "", "volume", "inspect", volumeName, "--format", "json")
	if err != nil {
		return "", err
	}

	path := volumeInfos[0].Mountpoint
	relToDocker, err := filepath.Rel("/var/lib/docker", path)
	if err != nil {
		return "", err
	}
	path = filepath.Join(rn.Sandbox.SandboxDir, "docker", relToDocker)

	return path, nil
}

func (rn *BoxSpecRunner) reconcileDockerVolumes(ctx context.Context) error {
	rn.volumeSpecHashes = nil
	for _, vol := range rn.BoxSpec.Volumes {
		h := sha256.New()
		err := json.NewEncoder(h).Encode(vol)
		if err != nil {
			return err
		}
		hash := hex.EncodeToString(h.Sum(nil))
		rn.volumeSpecHashes = append(rn.volumeSpecHashes, hash)

		volumePath, err := rn.createDockerVolume(ctx, vol.Name, hash)
		if err != nil {
			return err
		}

		err = rn.writeFileBundle(vol, volumePath)
		if err != nil {
			return err
		}

		err = rn.fixVolumePermissions(vol, volumePath)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rn *BoxSpecRunner) fixVolumePermissions(vol types.BoxVolumeSpec, volumePath string) error {
	err := os.Chown(volumePath, int(vol.RootUid), int(vol.RootGid))
	if err != nil {
		return err
	}
	rootMode, err := parseMode(vol.RootMode)
	if err != nil {
		return fmt.Errorf("failed to parse root dir mode: %w", err)
	}
	if rootMode != 0 {
		err = os.Chmod(volumePath, rootMode)
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
