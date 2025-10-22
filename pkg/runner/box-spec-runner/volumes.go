package box_spec_runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	ctypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/dboxed/dboxed/pkg/volume/volume_serve"
	"github.com/moby/sys/mountinfo"
)

func (rn *BoxSpecRunner) getVolumeWorkDir(uuid string) string {
	return filepath.Join(rn.WorkDir, "volumes", uuid)
}

func (rn *BoxSpecRunner) getVolumeMountDir(uuid string) string {
	return filepath.Join(rn.getVolumeWorkDir(uuid), "mount")
}

func (rn *BoxSpecRunner) reconcileVolumes(ctx context.Context, composeProjects map[string]*ctypes.Project, newVolumes []boxspec.DboxedVolume, allowDownService bool) error {
	oldVolumesByName := map[int64]*volume_serve.VolumeState{}
	newVolumeByName := map[int64]*boxspec.DboxedVolume{}

	oldVolumes, err := volume_serve.ListVolumeState(rn.getVolumeWorkDir(""))
	if err != nil {
		return err
	}

	for _, v := range oldVolumes {
		oldVolumesByName[v.Volume.ID] = v
	}
	for _, v := range newVolumes {
		newVolumeByName[v.Id] = &v
	}

	needDown := false
	var deleteVolumes []*volume_serve.VolumeState

	var createVolumes []*boxspec.DboxedVolume
	var mountVolumes []*boxspec.DboxedVolume
	var installServices []*boxspec.DboxedVolume

	for _, oldVolume := range oldVolumesByName {
		if _, ok := newVolumeByName[oldVolume.Volume.ID]; !ok {
			if allowDownService {
				rn.Log.InfoContext(ctx, "need to down services due to volume being deleted", slog.Any("volumeName", oldVolume.Volume.Name))
			}
			needDown = true
			deleteVolumes = append(deleteVolumes, oldVolume)
		}
	}
	for _, newVolume := range newVolumeByName {
		if oldVolume, ok := oldVolumesByName[newVolume.Id]; ok {
			mountDir := rn.getVolumeMountDir(oldVolume.Volume.Uuid)

			mounted, err := mountinfo.Mounted(mountDir)
			if err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			}
			if !mounted {
				mountVolumes = append(mountVolumes, newVolume)
				installServices = append(installServices, newVolume)
			}
		} else {
			createVolumes = append(createVolumes, newVolume)
			installServices = append(installServices, newVolume)
		}
	}
	if allowDownService && needDown {
		err = rn.runComposeDown(ctx, composeProjects, false, false)
		if err != nil {
			return err
		}
	}

	for _, v := range deleteVolumes {
		err = rn.uninstallVolumeService(ctx, v)
		if err != nil {
			return err
		}
		err = rn.releaseVolume(ctx, v)
		if err != nil {
			return err
		}
	}
	for _, v := range createVolumes {
		err = rn.createVolume(ctx, v)
		if err != nil {
			return err
		}
	}
	for _, v := range mountVolumes {
		err = rn.mountVolume(ctx, v)
		if err != nil {
			return err
		}
	}
	for _, v := range newVolumes {
		mountDir := rn.getVolumeMountDir(v.Uuid)
		err = rn.fixVolumePermissions(v, mountDir)
		if err != nil {
			return err
		}
	}
	for _, v := range installServices {
		err = rn.installVolumeService(ctx, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rn *BoxSpecRunner) createVolume(ctx context.Context, vol *boxspec.DboxedVolume) error {
	rn.Log.InfoContext(ctx, "creating volume-mount",
		slog.Any("name", vol.Name),
	)

	args := []string{
		"--work-dir", rn.WorkDir,
		"volume-mount",
		"create",
		vol.Uuid,
		"--box", rn.BoxSpec.Uuid,
	}

	err := rn.runDboxedVolume(ctx, args)
	if err != nil {
		return err
	}

	return nil
}

func (rn *BoxSpecRunner) mountVolume(ctx context.Context, vol *boxspec.DboxedVolume) error {
	rn.Log.InfoContext(ctx, "mounting volume",
		slog.Any("name", vol.Name),
	)

	args := []string{
		"--work-dir", rn.WorkDir,
		"volume-mount",
		"mount",
		vol.Uuid,
	}

	err := rn.runDboxedVolume(ctx, args)
	if err != nil {
		return err
	}

	return nil
}

func (rn *BoxSpecRunner) releaseVolume(ctx context.Context, vol *volume_serve.VolumeState) error {
	rn.Log.InfoContext(ctx, "releasing volume",
		slog.Any("name", vol.Volume.Name),
	)

	args := []string{
		"--work-dir", rn.WorkDir,
		"volume-mount",
		"release",
		vol.Volume.Uuid,
	}

	err := rn.runDboxedVolume(ctx, args)
	if err != nil {
		return err
	}
	return nil
}
func (rn *BoxSpecRunner) installVolumeService(ctx context.Context, vol *boxspec.DboxedVolume) error {
	rn.Log.InfoContext(ctx, "installing volume service",
		slog.Any("name", vol.Name),
	)

	args := []string{
		"--work-dir", rn.WorkDir,
		"volume-mount",
		"service",
		"install",
		vol.Uuid,
		"--backup-interval", vol.BackupInterval,
	}

	err := rn.runDboxedVolume(ctx, args)
	if err != nil {
		return err
	}
	return nil
}

func (rn *BoxSpecRunner) uninstallVolumeService(ctx context.Context, vol *volume_serve.VolumeState) error {
	rn.Log.InfoContext(ctx, "uninstalling volume service",
		slog.Any("name", vol.Volume.Name),
	)

	args := []string{
		"--work-dir", rn.WorkDir,
		"volume-mount",
		"service",
		"uninstall",
		vol.Volume.Uuid,
	}

	err := rn.runDboxedVolume(ctx, args)
	if err != nil {
		return err
	}
	return nil
}

func (rn *BoxSpecRunner) fixVolumePermissions(vol boxspec.DboxedVolume, mountDir string) error {
	st, err := os.Stat(mountDir)
	if err != nil {
		return err
	}

	newMode, err := rn.parseMode(vol.RootMode)
	if err != nil {
		return err
	}
	if st.Mode().Perm() != newMode {
		err = os.Chmod(mountDir, newMode)
		if err != nil {
			return err
		}
	}
	st2, ok := st.Sys().(*syscall.Stat_t)
	if ok {
		if st2.Uid != vol.RootUid || st2.Gid != vol.RootUid {
			err = os.Chown(mountDir, int(vol.RootUid), int(vol.RootGid))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (rn *BoxSpecRunner) parseMode(s string) (os.FileMode, error) {
	n, err := strconv.ParseInt(s, 8, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid file mode %s: %w", s, err)
	}
	fm := os.FileMode(n)
	if fm & ^os.ModePerm != 0 {
		return 0, fmt.Errorf("invalid file mode %s", s)
	}
	return fm, nil
}

func (rn *BoxSpecRunner) runDboxedVolume(ctx context.Context, args []string) error {
	c := util.CommandHelper{
		Command: "dboxed",
		Args:    args,
		Logger:  rn.Log,
		LogCmd:  true,
	}
	err := c.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}
