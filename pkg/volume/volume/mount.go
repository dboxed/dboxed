package volume

import (
	"context"
	"log/slog"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/util"
	"github.com/moby/sys/mountinfo"
)

func (v *Volume) Mount(ctx context.Context, mountTarget string, readOnly bool) error {
	lvDev, err := v.DevPath(true)
	if err != nil {
		return err
	}

	mounts, err := mountinfo.GetMounts(nil)
	if err != nil {
		return err
	}

	for _, m := range mounts {
		source, err := filepath.EvalSymlinks(m.Source)
		if err == nil {
			if m.Mountpoint == mountTarget && source == lvDev {
				slog.Info("volume already mounted", slog.Any("mountPoint", m.Mountpoint), slog.Any("source", m.Source))
				return nil
			}
		}
	}

	var args []string
	if readOnly {
		args = append(args, "-o", "ro")
	}

	args = append(args, lvDev, mountTarget)

	err = util.RunCommand(ctx, "mount", args...)
	if err != nil {
		return err
	}

	return nil
}

func (v *Volume) RemountReadOnly(ctx context.Context, mountTarget string) error {
	err := util.RunCommand(ctx, "mount", "-o", "remount,ro", mountTarget)
	if err != nil {
		return err
	}
	return nil
}

func (v *Volume) Unmount(ctx context.Context, mountTarget string) error {
	err := util.RunCommand(ctx, "umount", mountTarget)
	if err != nil {
		return err
	}
	return nil
}
