package volume

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/dboxed/dboxed/pkg/volume/lvm"
)

var AllowedFsTypes = []string{
	"ext2",
	"ext3",
	"ext4",
	"xfs",
	"btrfs",
}

type Volume struct {
	image string

	filesystemLv *lvm.LVEntry
}

func Open(ctx context.Context, image string, lockId string) (*Volume, error) {
	_, loHandle, err := GetOrAttachLoopDev(image, lockId)
	if err != nil {
		return nil, err
	}
	defer loHandle.Close()

	tag := fmt.Sprintf("dboxed-volume-lock-%s", lockId)

	lvs, err := lvm.FindLVsWithTag(ctx, tag)
	if err != nil {
		return nil, err
	}

	var filesystemLv *lvm.LVEntry
	for _, lv := range lvs {
		if slices.Contains(lv.LvTags.L, "filesystem") {
			filesystemLv = &lv
		}
	}
	if filesystemLv == nil {
		return nil, fmt.Errorf("logical volume with filesystem tag not found")
	}

	// this will cause device-mapper to open the loop dev and keep it open even though we detach
	err = lvm.LVActivate(ctx, filesystemLv.VgName, filesystemLv.LvName, true)
	if err != nil {
		return nil, err
	}

	v := &Volume{
		image:        image,
		filesystemLv: filesystemLv,
	}

	return v, nil
}

func (v *Volume) DevPath(evalSymlinks bool) (string, error) {
	return lvm.BuildDevPath(v.filesystemLv.VgName, v.filesystemLv.LvName, evalSymlinks)
}

func (v *Volume) SnapshotDevPath(snapshotName string, evalSymlinks bool) (string, error) {
	return lvm.BuildDevPath(v.filesystemLv.VgName, snapshotName, evalSymlinks)
}

func (v *Volume) Deactivate(ctx context.Context) error {
	return DeactivateVolume(ctx, v.filesystemLv.VgName)
}

func DeactivateVolume(ctx context.Context, vgName string) error {
	slog.Info("deactivating volume group", slog.Any("vgName", vgName))
	err := lvm.VGDeactivate(ctx, vgName)
	if err != nil {
		return err
	}

	// some hidden/internal LVs might still be active, so we explicitly deactivate them
	lvs, err := lvm.ListLVs(ctx)
	for _, lv := range lvs {
		if lv.VgName == vgName && lv.LvActive == "active" {
			slog.Info("deactivating internal/hidden volume", slog.Any("vgName", vgName), slog.Any("lvName", lv.LvName))
			err = lvm.LVActivateFullName(ctx, lv.LvFullName, false)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
