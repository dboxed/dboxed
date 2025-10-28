package volume

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/dboxed/dboxed/pkg/util"
	"github.com/dboxed/dboxed/pkg/volume/lvm"
	"github.com/dboxed/dboxed/pkg/volume/mount"
)

func (v *Volume) CreateSnapshot(ctx context.Context, snapshotName string, overwrite bool) error {
	snapLv, err := lvm.LVGet(ctx, v.filesystemLv.VgName, snapshotName)
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
		err = lvm.LVRemove(ctx, v.filesystemLv.VgName, snapshotName)
		if err != nil {
			return err
		}
	}

	_ = util.RunCommand(ctx, "sync")

	slog.Info("creating snapshot", slog.Any("snapshotName", snapshotName))
	err = lvm.LVSnapCreate100Free(ctx, v.filesystemLv.VgName, v.filesystemLv.LvName, snapshotName)
	if err != nil {
		return err
	}

	deferRemoveSnapshot := true
	defer func() {
		if deferRemoveSnapshot {
			err := lvm.LVRemove(ctx, v.filesystemLv.VgName, snapshotName)
			if err != nil {
				slog.Error("remove snapshot failed in defer", slog.Any("error", err))
			}
		}
	}()

	err = lvm.LVActivate(ctx, v.filesystemLv.VgName, snapshotName, true)
	if err != nil {
		return err
	}

	deferRemoveSnapshot = false
	return nil
}

func (v *Volume) DeleteSnapshot(ctx context.Context, snapshotName string) error {
	return lvm.LVRemove(ctx, v.filesystemLv.VgName, snapshotName)
}

func (v *Volume) MountSnapshot(ctx context.Context, snapshotName string, mountTarget string) error {
	lvDev, err := lvm.BuildDevPath(v.filesystemLv.VgName, snapshotName, false)
	if err != nil {
		return err
	}
	err = mount.Mount(ctx, "", lvDev, mountTarget, true)
	if err != nil {
		return err
	}
	return nil
}

func (v *Volume) UnmountSnapshot(ctx context.Context, snapshotName string) error {
	isMounted, err := v.IsSnapshotMounted(snapshotName)
	if err != nil {
		return err
	}
	if !isMounted {
		return nil
	}

	lvDev, err := lvm.BuildDevPath(v.filesystemLv.VgName, snapshotName, false)
	if err != nil {
		return err
	}
	return mount.Unmount(ctx, lvDev)
}

func (v *Volume) IsSnapshotMounted(snapshotName string) (bool, error) {
	lvDev, err := lvm.BuildDevPath(v.filesystemLv.VgName, snapshotName, true)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return mount.IsMountedSource(lvDev)
}
