package box_spec_runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/dboxed/dboxed-common/util"
	"github.com/dboxed/dboxed/pkg/types"
)

type volumeInterfaceFileBundle struct {
}

func (vi volumeInterfaceFileBundle) WorkDirBase(vol types.BoxVolumeSpec) string {
	return fmt.Sprintf("file-bundle-%s", vol.Name)
}

func (vi volumeInterfaceFileBundle) IsReadOnly(vol types.BoxVolumeSpec) bool {
	return true
}

func (vi volumeInterfaceFileBundle) Create(ctx context.Context, vol types.BoxVolumeSpec) error {
	workDir := getVolumeWorkDir(vi, vol)
	mountDir := getVolumeMountDir(vi, vol)

	slog.InfoContext(ctx, "creating file bundle volume",
		slog.Any("name", vol.Name),
		slog.Any("workDir", workDir),
		slog.Any("mountDir", mountDir),
	)

	err := os.MkdirAll(workDir, 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(mountDir, 0700)
	if err != nil {
		return err
	}

	err = vi.createFileBundle(ctx, vol, mountDir)
	if err != nil {
		return err
	}

	err = fixVolumePermissions(vol, mountDir)
	if err != nil {
		return err
	}
	return nil
}

func (vi volumeInterfaceFileBundle) Delete(ctx context.Context, vol types.BoxVolumeSpec) error {
	workDir := getVolumeWorkDir(vi, vol)

	slog.InfoContext(ctx, "deleting file bundle volume",
		slog.Any("name", vol.Name),
		slog.Any("workDir", workDir),
	)

	err := os.RemoveAll(workDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (vi volumeInterfaceFileBundle) CheckRecreateNeeded(oldVol types.BoxVolumeSpec, newVol types.BoxVolumeSpec) bool {
	return util.MustJson(oldVol) != util.MustJson(newVol)
}

func (vi volumeInterfaceFileBundle) createFileBundle(ctx context.Context, vol types.BoxVolumeSpec, volumeDir string) error {
	fb := vol.FileBundle

	for _, f := range fb.Files {
		err := vi.writeFileBundleEntry(volumeDir, f)
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

		p, err := securejoin.SecureJoin(volumeDir, f.Path)
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

	return nil
}

func (vi volumeInterfaceFileBundle) writeFileBundleEntry(bundlePath string, f types.FileBundleEntry) error {
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
	case types.FileBundleEntryFile, "":
		err = util.AtomicWriteFile(p, d, fileMode.Perm())
		if err != nil {
			return err
		}
	case types.FileBundleEntryDir:
		err = os.Mkdir(p, fileMode.Perm())
		if err != nil {
			if !os.IsExist(err) {
				return err
			}
		}
	case types.FileBundleEntrySymlink:
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
