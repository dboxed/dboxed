package mount

import (
	"context"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/util/command_helper"
	"github.com/moby/sys/mountinfo"
)

func ListMounts() ([]*mountinfo.Info, error) {
	ret, err := mountinfo.GetMounts(nil)
	if err != nil {
		return nil, err
	}

	for i := range ret {
		p, err := filepath.EvalSymlinks(ret[i].Source)
		if err == nil {
			ret[i].Source = p
		}
		p, err = filepath.EvalSymlinks(ret[i].Mountpoint)
		if err == nil {
			ret[i].Mountpoint = p
		}
	}
	return ret, nil
}

func GetMountBySource(source string) (*mountinfo.Info, error) {
	mounts, err := ListMounts()
	if err != nil {
		return nil, err
	}

	source, err = filepath.EvalSymlinks(source)
	if err != nil {
		return nil, err
	}

	for _, m := range mounts {
		if m.Source == source {
			return m, nil
		}
	}
	return nil, os.ErrNotExist
}

func GetMountByMountpoint(mountPoint string) (*mountinfo.Info, error) {
	mounts, err := ListMounts()
	if err != nil {
		return nil, err
	}

	mountPoint, err = filepath.EvalSymlinks(mountPoint)
	if err != nil {
		return nil, err
	}

	for _, m := range mounts {
		if m.Mountpoint == mountPoint {
			return m, nil
		}
	}
	return nil, os.ErrNotExist
}

func IsMountedSource(source string) (bool, error) {
	_, err := GetMountBySource(source)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	} else {
		return true, nil
	}
}

func Mount(ctx context.Context, fsType string, source string, target string, readOnly bool) error {
	var args []string
	if fsType != "" {
		args = append(args, "-t", fsType)
	}
	if readOnly {
		args = append(args, "-o", "ro")
	}

	args = append(args, source, target)

	err := command_helper.RunCommand(ctx, "mount", args...)
	if err != nil {
		return err
	}
	return nil
}

func Unmount(ctx context.Context, target string) error {
	err := command_helper.RunCommand(ctx, "umount", target)
	if err != nil {
		return err
	}
	return nil
}
