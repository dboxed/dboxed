package volume

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/dboxed/dboxed/pkg/volume/mount"
)

const refFile = ".dboxed-loop-ref"

const RefPrefix = "dboxed-volume"

func BuildRef(lockId string) string {
	h := strings.ReplaceAll(lockId, "-", "")
	s := fmt.Sprintf("%s-%s", RefPrefix, h)
	if len(s) > 64 {
		s = s[:64]
	}
	return s
}

func WriteLoopRef(ctx context.Context, refMountDir string, lockId string) error {
	m, err := mount.GetMountByMountpoint(refMountDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	if m != nil {
		if m.FSType != "tmpfs" {
			return fmt.Errorf("unexpected filesystem type %s for loop-ref", m.FSType)
		}
	} else {
		err = os.MkdirAll(refMountDir, 0700)
		if err != nil {
			return err
		}
		slog.Info("mounting tmpfs to hold loop-ref")
		err = mount.Mount(ctx, "tmpfs", "none", refMountDir, false)
		if err != nil {
			return err
		}
	}

	err = os.WriteFile(filepath.Join(refMountDir, refFile), []byte(BuildRef(lockId)), 0644)
	if err != nil {
		return err
	}
	return nil
}

func UnmountLoopRefs(ctx context.Context, refMountDir string) error {
	m, err := mount.GetMountByMountpoint(refMountDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	if m.FSType != "tmpfs" {
		return fmt.Errorf("unexpected filesystem type %s for loop-ref", m.FSType)
	}

	return mount.Unmount(ctx, refMountDir)
}

func ReadLoopRef(refMountDir string) (string, error) {
	b, err := os.ReadFile(filepath.Join(refMountDir, refFile))
	if err != nil {
		return "", err
	}
	return string(b), nil
}
