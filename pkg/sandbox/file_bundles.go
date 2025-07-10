package sandbox

import (
	"context"
	"fmt"
	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/koobox/unboxed/pkg/types"
	"os"
	"path/filepath"
)

func (rn *Sandbox) getBaseBundlesPathOnHost() string {
	return filepath.Join(rn.SandboxDir, "bundles")
}

func (rn *Sandbox) getBundlePathOnHost(name string) string {
	return filepath.Join(rn.getBaseBundlesPathOnHost(), name)
}

func (rn *Sandbox) writeFileBundles(ctx context.Context) error {
	err := os.RemoveAll(rn.getBaseBundlesPathOnHost())
	if err != nil {
		return err
	}
	err = os.MkdirAll(rn.getBaseBundlesPathOnHost(), 0700)
	if err != nil {
		return err
	}

	for _, fb := range rn.BoxSpec.FileBundles {
		err := rn.writeFileBundle(fb)
		if err != nil {
			return err
		}
	}
	return nil
}

func (rn *Sandbox) writeFileBundle(fb types.FileBundle) error {
	bundlePath := rn.getBundlePathOnHost(fb.Name)

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

	return nil
}

func (rn *Sandbox) writeFileBundleEntry(bundlePath string, f types.FileBundleEntry) error {
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
