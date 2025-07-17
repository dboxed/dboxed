package run_infra_sandbox

import (
	"context"
	"fmt"
	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/koobox/unboxed/pkg/types"
	"log/slog"
	"os"
	"path/filepath"
)

func (rn *RunInfraSandbox) getBundleVolumeName(name string) string {
	return fmt.Sprintf("unboxed-bundle-%s", name)
}

func (rn *RunInfraSandbox) createBundleVolume(ctx context.Context, name string) (string, error) {
	volumeName := rn.getBundleVolumeName(name)

	slog.InfoContext(ctx, "creating bundle volumes", slog.Any("volumeName", volumeName))
	_, err := rn.runDockerCli(ctx, true, "volume", "create", volumeName)
	if err != nil {
		return "", err
	}

	var volumeInfos []types.DockerVolume
	err = rn.runDockerCliJson(ctx, &volumeInfos, "volume", "inspect", volumeName, "--format", "json")
	if err != nil {
		return "", err
	}
	return volumeInfos[0].Mountpoint, nil
}

func (rn *RunInfraSandbox) createBundleVolumes(ctx context.Context) error {
	for _, fb := range rn.conf.BoxSpec.FileBundles {
		volumePath, err := rn.createBundleVolume(ctx, fb.Name)
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

func (rn *RunInfraSandbox) writeFileBundle(fb types.FileBundle, bundlePath string) error {
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
		fileMode := os.FileMode(f.Mode)

		p, err := securejoin.SecureJoin(bundlePath, f.Path)
		if err != nil {
			return err
		}
		err = os.Chmod(p, fileMode.Perm())
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
	if fb.RootMode != 0 {
		if os.FileMode(fb.RootMode) & ^os.ModePerm != 0 {
			return fmt.Errorf("not allowed mode %o", fb.RootMode)
		}
		err = os.Chmod(bundlePath, os.FileMode(fb.RootMode))
		if err != nil {
			return err
		}
	}

	return nil
}

func (rn *RunInfraSandbox) writeFileBundleEntry(bundlePath string, f types.FileBundleEntry) error {
	fileMode := os.FileMode(f.Mode)
	if fileMode & ^types.AllowedModeMask != 0 {
		return fmt.Errorf("invalid file mode %o for entry '%s'", fileMode, f.Path)
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
	if fileMode.IsDir() {
		err = os.Mkdir(p, fileMode.Perm())
		if err != nil {
			return err
		}
	} else if fileMode.IsRegular() {
		err = os.WriteFile(p, d, fileMode.Perm())
		if err != nil {
			return err
		}
	} else if fileMode.Type() == os.ModeSymlink {
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
	}

	return nil
}
