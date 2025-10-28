//go:build linux

package volume_mount

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/volume/lvm"
	"github.com/dboxed/dboxed/pkg/volume/mount"
	"github.com/dboxed/dboxed/pkg/volume/volume_serve"
)

type ForceRemoveCmd struct {
	Volume string `help:"Specify volume" required:"" arg:""`
}

func (cmd *ForceRemoveCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	baseDir := filepath.Join(g.WorkDir, "volumes")
	volumeState, err := commandutils.GetMountedVolume(baseDir, cmd.Volume)
	if err != nil {
		return err
	}
	dir := filepath.Join(baseDir, volumeState.MountName)

	vs, err := volume_serve.LoadVolumeState(dir)
	if err != nil {
		return err
	}
	c, err := baseclient.New(nil, vs.ClientAuth, false)
	if err != nil {
		slog.Warn("failed to create client", "error", err)
	}

	slog.Info("force-removing volume mount - no backup will be performed", slog.Any("volume", volumeState.Volume.Name), slog.Any("mountName", volumeState.MountName))

	if c != nil && vs.Volume != nil && vs.Volume.LockId != nil {
		slog.Info("trying to release volume lock")
		c2 := clients.VolumesClient{Client: c}
		_, err = c2.VolumeRelease(ctx, vs.Volume.ID, models.VolumeReleaseRequest{
			LockId: *vs.Volume.LockId,
		})
		if err != nil {
			slog.Warn("failed to release volume lock", "error", err)
		}
	}

	if vs.Volume != nil {
		tag := fmt.Sprintf("dboxed-volume-lock-%s", *vs.Volume.LockId)
		lvs, err := lvm.FindLVsWithTag(ctx, tag)
		if err != nil {
			return err
		}

		vgs := map[string]struct{}{}

		mounts, err := mount.ListMounts()
		if err != nil {
			return err
		}
		for _, lv := range lvs {
			vgs[lv.VgName] = struct{}{}

			devPath, err := lvm.BuildDevPath(lv.VgName, lv.LvName, true)
			if err != nil {
				return err
			}
			for _, m := range mounts {
				if m.Source == devPath {
					slog.Info("trying to unmount logical volume", "source", m.Source, "devPath", devPath)
					err = mount.Unmount(ctx, m.Mountpoint)
					if err != nil {
						return err
					}
				}
			}
		}
		for vgName := range vgs {
			slog.Info("trying to deactivate logical volume group", "vgName", vgName)
			err = lvm.VGDeactivate(ctx, vgName)
			if err != nil {
				return err
			}
		}

		loopRefDir := filepath.Join(dir, "loop-ref")
		for _, m := range mounts {
			if m.Mountpoint == loopRefDir {
				slog.Info("trying to unmount loop-ref dir", "loopRefDir", loopRefDir)
				err = mount.Unmount(ctx, m.Mountpoint)
				if err != nil {
					return err
				}
				break
			}
		}
	}

	slog.Info("force removing volume directory")
	err = os.RemoveAll(dir)
	if err != nil {
		return fmt.Errorf("failed to remove volume directory: %w", err)
	}

	slog.Info("force-remove completed successfully")

	return nil
}
