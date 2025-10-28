package volume

import (
	"context"
	"log/slog"
	"os"

	"github.com/dboxed/dboxed/pkg/util"
	"github.com/dboxed/dboxed/pkg/volume/mount"
)

func (v *Volume) Mount(ctx context.Context, mountTarget string, readOnly bool) error {
	lvDev, err := v.DevPath(true)
	if err != nil {
		return err
	}

	m, err := mount.GetMountBySource(lvDev)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	if m != nil {
		slog.Info("volume already mounted", slog.Any("mountPoint", m.Mountpoint), slog.Any("source", m.Source))
		return nil
	}

	err = mount.Mount(ctx, "", lvDev, mountTarget, readOnly)
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
	return mount.Unmount(ctx, mountTarget)
}

func (v *Volume) IsMounted() (bool, error) {
	lvDev, err := v.DevPath(true)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return mount.IsMountedSource(lvDev)
}
