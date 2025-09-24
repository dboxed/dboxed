package volume

import (
	"fmt"
	"log/slog"
	"path/filepath"
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

	loopDev      string
	filesystemLv *lvm.LVEntry
}

func Open(image string, lockId string) (*Volume, error) {
	loDev, loHandle, err := GetOrAttachLoopDev(image, lockId)
	if err != nil {
		return nil, err
	}
	defer loHandle.Close()

	lvs, err := lvm.FindPVLVs(loDev.Path())
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
	err = lvm.LVActivate(filesystemLv.VgName, filesystemLv.LvName, true)
	if err != nil {
		return nil, err
	}

	v := &Volume{
		image:        image,
		loopDev:      loDev.Path(),
		filesystemLv: filesystemLv,
	}

	return v, nil
}

func (v *Volume) DevPath(evalSymlinks bool) (string, error) {
	return buildDevPath(v.filesystemLv.VgName, v.filesystemLv.LvName, evalSymlinks)
}

func (v *Volume) SnapshotDevPath(snapshotName string, evalSymlinks bool) (string, error) {
	return buildDevPath(v.filesystemLv.VgName, snapshotName, evalSymlinks)
}

func (v *Volume) Deactivate() error {
	return DeactivateVolume(v.filesystemLv.VgName)
}

func DeactivateVolume(vgName string) error {
	slog.Info("deactivating volume group", slog.Any("vgName", vgName))
	err := lvm.VGDeactivate(vgName)
	if err != nil {
		return err
	}

	// some hidden/internal LVs might still be active, so we explicitly deactivate them
	lvs, err := lvm.ListLVs()
	for _, lv := range lvs {
		if lv.VgName == vgName && lv.LvActive == "active" {
			slog.Info("deactivating internal/hidden volume", slog.Any("vgName", vgName), slog.Any("lvName", lv.LvName))
			err = lvm.LVActivateFullName(lv.LvFullName, false)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func buildDevPath(vgName string, lvName string, evalSymlinks bool) (string, error) {
	p := filepath.Join("/dev", vgName, lvName)
	if evalSymlinks {
		var err error
		p, err = filepath.EvalSymlinks(p)
		if err != nil {
			return "", err
		}
	}
	return p, nil
}
