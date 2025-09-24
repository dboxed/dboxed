package volume

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/dboxed/dboxed/pkg/util"
	"github.com/moby/sys/mountinfo"
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

func WriteLoopRef(refMountDir string, lockId string) error {
	mounts, err := mountinfo.GetMounts(nil)
	if err != nil {
		return err
	}

	found := false
	for _, m := range mounts {
		if m.Mountpoint == refMountDir && m.FSType == "tmpfs" {
			found = true
			break
		}
	}
	if !found {
		err = os.MkdirAll(refMountDir, 0700)
		if err != nil {
			return err
		}
		slog.Info("mounting tmpfs to hold loop-ref")
		err = util.RunCommand("mount", "-t", "tmpfs", "none", refMountDir)
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

func UnmountLoopRefs(refMountDir string) error {
	mounts, err := mountinfo.GetMounts(nil)
	if err != nil {
		return err
	}

	found := false
	for _, m := range mounts {
		if m.Mountpoint == refMountDir && m.FSType == "tmpfs" {
			found = true
			break
		}
	}
	if !found {
		return nil
	}

	err = util.RunCommand("umount", refMountDir)
	if err != nil {
		return err
	}
	return nil
}

func ReadLoopRef(refMountDir string) (string, error) {
	b, err := os.ReadFile(filepath.Join(refMountDir, refFile))
	if err != nil {
		return "", err
	}
	return string(b), nil
}
