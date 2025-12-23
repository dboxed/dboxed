package box_spec_runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	ctypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/runner/compose"
	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/dboxed/dboxed/pkg/util/command_helper"
	"github.com/dboxed/dboxed/pkg/version"
	"github.com/dboxed/dboxed/pkg/volume/volume_serve"
)

func (rn *BoxSpecRunner) getVolumeWorkDir(id string) string {
	return filepath.Join(consts.DboxedDataDir, "volumes", id)
}

func (rn *BoxSpecRunner) getDboxedVolumeMountDir(id string) string {
	return filepath.Join(rn.getVolumeWorkDir(id), "mount")
}

func (rn *BoxSpecRunner) getContentFilePath(target string) string {
	h := util.Sha256Sum([]byte(target))
	return filepath.Join(consts.DboxedDataDir, "content-volumes", h)
}

func (rn *BoxSpecRunner) updateServiceVolume(volume *ctypes.ServiceVolumeConfig) error {
	if volume.Type == "dboxed" {
		dv := rn.BoxSpec.GetVolumeByName(volume.Source)
		if dv == nil {
			return fmt.Errorf("dboxed volume with name %s not found", volume.Source)
		}
		volume.Type = ctypes.VolumeTypeBind
		volume.Source = rn.getDboxedVolumeMountDir(dv.ID)
		return nil
	} else if volume.Type == "content" {
		volume.Type = ctypes.VolumeTypeBind
		volume.Source = rn.getContentFilePath(volume.Target)
		volume.ReadOnly = true
		return nil
	} else {
		return nil
	}
}

func (rn *BoxSpecRunner) reconcileContentVolumes(ctx context.Context) error {
	_, composeProjects, err := rn.loadBoxSpecComposeProjects(ctx)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Join(consts.DboxedDataDir, "content-volumes"), 0700)
	if err != nil {
		return err
	}

	for _, cp := range composeProjects {
		for _, s := range cp.Project.Services {
			for _, v := range s.Volumes {
				if v.Type == "content" {
					pth := rn.getContentFilePath(v.Target)
					var content string
					ok, err := v.Extensions.Get("x-content", &content)
					if err != nil {
						return fmt.Errorf("error retrieving content from service volume for target %s: %w", v.Target, err)
					}
					if !ok {
						return fmt.Errorf("missing content in service volume for target %s", v.Target)
					}
					err = os.WriteFile(pth, []byte(content), 0400)
					if err != nil {
						return fmt.Errorf("failed writing volume content for target %s: %w", v.Target, err)
					}
				}
			}
		}
	}
	return nil
}

func (rn *BoxSpecRunner) reconcileDboxedVolumes(ctx context.Context, allowDownService bool) error {
	oldVolumesByName := map[string]*volume_serve.VolumeState{}
	newVolumeByName := map[string]*boxspec.DboxedVolume{}

	oldVolumes, err := volume_serve.ListVolumeState(rn.getVolumeWorkDir(""))
	if err != nil {
		return err
	}

	for _, v := range oldVolumes {
		oldVolumesByName[v.Volume.ID] = v
	}
	for _, v := range rn.BoxSpec.Volumes {
		newVolumeByName[v.ID] = &v
	}

	needDown := false
	p, err := rn.buildDboxedVolumesComposeProject(ctx)
	if err != nil {
		return err
	}
	ch := compose.ComposeHelper{
		BaseDir: rn.composeBaseDir,
		Project: p,
	}

	for _, oldVolume := range oldVolumesByName {
		if _, ok := newVolumeByName[oldVolume.Volume.ID]; !ok {
			if allowDownService {
				slog.InfoContext(ctx, "need to down services due to volume being deleted", slog.Any("volumeName", oldVolume.Volume.Name))
			}
			needDown = true
		}
	}
	if !needDown && p != nil {
		changed, err := ch.CheckRecreateNeeded(ctx)
		if err != nil {
			return err
		}
		if changed {
			if allowDownService {
				slog.InfoContext(ctx, "need to down services due to volume needed to be recreated")
			}
			needDown = true
		}
	}

	if allowDownService && needDown {
		composeProjects, _, err := rn.loadBoxSpecComposeProjects(ctx)
		for name, _ := range composeProjects {
			err = compose.RunComposeDown(ctx, name, false, false)
			if err != nil {
				return err
			}
		}
	}

	if p != nil {
		err = ch.RunPull(ctx)
		if err != nil {
			return err
		}
		err = ch.RunUp(ctx, true)
		if err != nil {
			return err
		}
	} else {
		err = rn.downDboxedVolumes(ctx)
		if err != nil {
			return err
		}
	}

	for _, v := range rn.BoxSpec.Volumes {
		mountDir := rn.getDboxedVolumeMountDir(v.ID)
		err = rn.fixVolumePermissions(v, mountDir)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rn *BoxSpecRunner) downDboxedVolumes(ctx context.Context) error {
	slog.InfoContext(ctx, "downing dboxed volumes")

	err := compose.RunComposeDown(ctx, "dboxed-volumes", false, false)
	if err != nil {
		return err
	}

	return nil
}

func (rn *BoxSpecRunner) buildDboxedVolumesComposeProject(ctx context.Context) (*ctypes.Project, error) {
	if len(rn.BoxSpec.Volumes) == 0 {
		return nil, nil
	}

	p := &ctypes.Project{
		Name:     "dboxed-volumes",
		Services: map[string]ctypes.ServiceConfig{},
		Volumes:  map[string]ctypes.VolumeConfig{},
	}

	dboxedBinVolume := ctypes.ServiceVolumeConfig{
		Type:   "bind",
		Source: "/usr/bin/dboxed",
		Target: "/usr/bin/dboxed",
	}

	clientAuth := rn.Client.GetClientAuth(true)

	for _, dv := range rn.BoxSpec.Volumes {
		p.Services[dv.Name] = ctypes.ServiceConfig{
			Image:       version.GetDefaultVolumeInfraImage(),
			Privileged:  true,
			NetworkMode: "host",
			Volumes: []ctypes.ServiceVolumeConfig{
				dboxedBinVolume,
				{
					Type:   "bind",
					Source: "/dev",
					Target: "/dev",
				},
				{
					Type:   "bind",
					Source: consts.VolumesDir,
					Target: consts.VolumesDir,
					Bind: &ctypes.ServiceVolumeBind{
						Propagation: "shared",
					},
				},
			},
			Environment: map[string]*string{
				"DBOXED_API_URL":   &clientAuth.ApiUrl,
				"DBOXED_API_TOKEN": clientAuth.StaticToken,
				"ASD":              util.Ptr("asd1"),
			},
			Entrypoint: []string{
				"dboxed",
				"volume-mount",
				"serve",
				dv.ID,
				"--backup-interval", dv.BackupInterval,
				"--ready-file", "/volume-ready",
			},
			HealthCheck: &ctypes.HealthCheckConfig{
				Test: []string{
					"CMD",
					"test",
					"-f",
					"/volume-ready",
				},
				Interval:    util.Ptr(ctypes.Duration(time.Second * 5)),
				StartPeriod: util.Ptr(ctypes.Duration(time.Minute * 10)),
			},
			StopGracePeriod: util.Ptr(ctypes.Duration(time.Minute * 10)),
			Init:            util.Ptr(true),
		}
	}

	return p, nil
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
	c := command_helper.CommandHelper{
		Command: "dboxed",
		Args:    args,
		LogCmd:  true,
	}
	err := c.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}
