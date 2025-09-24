package volume

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/util"
	"github.com/dboxed/dboxed/pkg/volume/lvm"
	"github.com/moby/sys/mountinfo"
)

func (v *Volume) CreateSnapshot(snapshotName string, overwrite bool) error {
	snapLv, err := lvm.LVGet(v.filesystemLv.VgName, snapshotName)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	if snapLv != nil {
		if !overwrite {
			return fmt.Errorf("snapshot %s already exists", snapshotName)
		}
		slog.Info("snapshot already exists, removing it", slog.Any("snapshotName", snapshotName))
		err = lvm.LVRemove(v.filesystemLv.VgName, snapshotName)
		if err != nil {
			return err
		}
	}

	_ = util.RunCommand("sync")

	slog.Info("creating snapshot", slog.Any("snapshotName", snapshotName))
	err = lvm.LVSnapCreate100Free(v.filesystemLv.VgName, v.filesystemLv.LvName, snapshotName)
	if err != nil {
		return err
	}

	deferRemoveSnapshot := true
	defer func() {
		if deferRemoveSnapshot {
			err := lvm.LVRemove(v.filesystemLv.VgName, snapshotName)
			if err != nil {
				slog.Error("remove snapshot failed in defer", slog.Any("error", err))
			}
		}
	}()

	err = lvm.LVActivate(v.filesystemLv.VgName, snapshotName, true)
	if err != nil {
		return err
	}

	deferRemoveSnapshot = false
	return nil
}

func (v *Volume) DeleteSnapshot(snapshotName string) error {
	return lvm.LVRemove(v.filesystemLv.VgName, snapshotName)
}

func (v *Volume) MountSnapshot(snapshotName string, mountTarget string) error {
	lvDev, err := buildDevPath(v.filesystemLv.VgName, snapshotName, false)
	if err != nil {
		return err
	}
	err = util.RunCommand("mount", "-oro", lvDev, mountTarget)
	if err != nil {
		return err
	}
	return nil
}

func (v *Volume) UnmountSnapshot(snapshotName string) error {
	isMounted, err := v.IsSnapshotMounted(snapshotName)
	if err != nil {
		return err
	}
	if !isMounted {
		return nil
	}

	lvDev, err := buildDevPath(v.filesystemLv.VgName, snapshotName, false)
	if err != nil {
		return err
	}
	err = util.RunCommand("umount", lvDev)
	return err
}

func (v *Volume) IsSnapshotMounted(snapshotName string) (bool, error) {
	mounts, err := mountinfo.GetMounts(nil)
	if err != nil {
		return false, err
	}
	lvDev, err := buildDevPath(v.filesystemLv.VgName, snapshotName, true)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	for _, m := range mounts {
		source, err := filepath.EvalSymlinks(m.Source)
		if err == nil {
			if source == lvDev {
				return true, nil
			}
		}
	}
	return false, nil
}
