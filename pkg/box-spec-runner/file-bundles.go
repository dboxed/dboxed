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

	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/dboxed/dboxed/pkg/types"
	"github.com/dboxed/dboxed/pkg/util"
)

func (rn *BoxSpecRunner) getBundleVolumeName(name string, hash string) string {
	return fmt.Sprintf("dboxed-bundle-%s-%s", name, hash[:6])
}

func (rn *BoxSpecRunner) createBundleVolume(ctx context.Context, name string, hash string) (string, error) {
	volumeName := rn.getBundleVolumeName(name, hash)

	slog.InfoContext(ctx, "creating bundle volumes", slog.Any("volumeName", volumeName))
	_, err := rn.Sandbox.RunDockerCli(ctx, true, "", "volume", "create", volumeName)
	if err != nil {
		return "", err
	}

	var volumeInfos []types.DockerVolume
	err = rn.Sandbox.RunDockerCliJson(ctx, &volumeInfos, "", "volume", "inspect", volumeName, "--format", "json")
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

func (rn *BoxSpecRunner) createBundleVolumes(ctx context.Context) error {
	rn.bundleContentHashes = nil
	for _, fb := range rn.BoxSpec.FileBundles {
		h := sha256.New()
		err := json.NewEncoder(h).Encode(fb)
		if err != nil {
			return err
		}
		hash := hex.EncodeToString(h.Sum(nil))
		rn.bundleContentHashes = append(rn.bundleContentHashes, hash)

		volumePath, err := rn.createBundleVolume(ctx, fb.Name, hash)
		if err != nil {
			return err
		}

		err = rn.writeFileBundle(fb, volumePath)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rn *BoxSpecRunner) writeFileBundle(fb types.FileBundle, bundlePath string) error {
	err := os.MkdirAll(bundlePath, 0700)
	if err != nil {
		return err
	}

	for _, f := range fb.Files {
		err := rn.writeFileBundleEntry(bundlePath, f)
		if err != nil {
			return err
		}
	}

	// now fix permissions
	for _, f := range fb.Files {
		fm, err := parseMode(f.Mode)
		if err != nil {
			return err
		}

		p, err := securejoin.SecureJoin(bundlePath, f.Path)
		if err != nil {
			return err
		}
		err = os.Chmod(p, fm)
		if err != nil {
			return err
		}

		err = os.Chown(p, f.Uid, f.Gid)
		if err != nil {
			return err
		}
	}

	err = os.Chown(bundlePath, int(fb.RootUid), int(fb.RootGid))
	if err != nil {
		return err
	}
	rootMode, err := parseMode(fb.RootMode)
	if err != nil {
		return fmt.Errorf("failed to parse root dir mode: %w", err)
	}
	if rootMode != 0 {
		err = os.Chmod(bundlePath, rootMode)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rn *BoxSpecRunner) writeFileBundleEntry(bundlePath string, f types.FileBundleEntry) error {
	fileMode, err := parseMode(f.Mode)
	if err != nil {
		return fmt.Errorf("failed to parse file mode for %s: %w", f.Path, err)
	}

	p, err := securejoin.SecureJoin(bundlePath, f.Path)
	if err != nil {
		return err
	}

	// Create parent dirs first. We'll later fix permissions of these
	err = os.MkdirAll(filepath.Dir(p), 0755)
	if err != nil {
		return err
	}

	d, err := f.GetDecodedData()
	if err != nil {
		return err
	}
	switch f.Type {
	case "file", "":
		err = util.AtomicWriteFile(p, d, fileMode.Perm())
		if err != nil {
			return err
		}
	case "dir":
		err = os.Mkdir(p, fileMode.Perm())
		if err != nil {
			return err
		}
	case "symlink":
		err = os.Symlink(string(d), p)
		if err != nil {
			return err
		}
		// verify that the symlink does not leave the bundle
		fh, err := securejoin.OpenInRoot(bundlePath, f.Path)
		if err != nil {
			return err
		}
		err = fh.Close()
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported file type %s for %s", f.Type, f.Path)
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
