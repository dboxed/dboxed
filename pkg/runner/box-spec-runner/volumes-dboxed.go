package box_spec_runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dboxed/dboxed-common/util"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/runner/consts"
)

type volumeInterfaceDboxed struct {
	rn *BoxSpecRunner
}

func (vi volumeInterfaceDboxed) WorkDirBase(vol boxspec.BoxVolumeSpec) string {
	apiUrlHash := util.Sha256Sum([]byte(vol.Dboxed.ApiUrl))
	return fmt.Sprintf("dboxed-volume-%s-%d-%d", apiUrlHash[:6], vol.Dboxed.RepositoryId, vol.Dboxed.VolumeId)
}

func (vi volumeInterfaceDboxed) IsReadOnly(vol boxspec.BoxVolumeSpec) bool {
	return false
}

func (vi volumeInterfaceDboxed) Create(ctx context.Context, vol boxspec.BoxVolumeSpec) error {
	workDir := getVolumeWorkDir(vi, vol)
	mountDir := getVolumeMountDir(vi, vol)

	slog.InfoContext(ctx, "creating dboxed-volume volume",
		slog.Any("name", vol.Name),
		slog.Any("workDir", workDir),
		slog.Any("mountDir", mountDir),
	)

	err := os.MkdirAll(workDir, 0700)
	if err != nil {
		return err
	}

	args := []string{
		"volume",
		"lock",
		"--api-url", vol.Dboxed.ApiUrl,
		"--repo", fmt.Sprintf("%d", vol.Dboxed.RepositoryId),
		"--volume", fmt.Sprintf("%d", vol.Dboxed.VolumeId),
		"--dir", workDir,
	}

	err = vi.runDboxedVolume(ctx, vol.Dboxed.Token, args)
	if err != nil {
		return err
	}

	err = fixVolumePermissions(vol, mountDir)
	if err != nil {
		return err
	}

	err = vi.createS6Service(ctx, vol)
	if err != nil {
		return err
	}

	return nil
}

func (vi volumeInterfaceDboxed) Delete(ctx context.Context, vol boxspec.BoxVolumeSpec) error {
	workDir := getVolumeWorkDir(vi, vol)

	slog.InfoContext(ctx, "deleting dboxed-volume volume",
		slog.Any("name", vol.Name),
		slog.Any("workDir", workDir),
	)

	err := vi.deleteS6Service(ctx, vol)
	if err != nil {
		return err
	}

	args := []string{
		"volume",
		"release",
		"--api-url", vol.Dboxed.ApiUrl,
		"--dir", workDir,
	}

	err = vi.runDboxedVolume(ctx, vol.Dboxed.Token, args)
	if err != nil {
		return err
	}

	err = os.RemoveAll(workDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (vi volumeInterfaceDboxed) createS6Service(ctx context.Context, vol boxspec.BoxVolumeSpec) error {
	name := vi.WorkDirBase(vol)
	workDir := getVolumeWorkDir(vi, vol)
	logDir := filepath.Join(consts.DboxedDataDir, "logs/s6", name)
	serviceDir := filepath.Join("/etc/services.d", name)
	serviceLogDir := filepath.Join(serviceDir, "log")
	symlinkPath := filepath.Join("/run/service", name)

	err := os.MkdirAll(logDir, 0700)
	if err != nil {
		return err
	}
	err = util.AtomicWriteFile(filepath.Join(logDir, "log-format"), []byte("slog-json"), 0644)
	if err != nil {
		return err
	}

	err = os.MkdirAll(serviceDir, 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(serviceLogDir, 0700)
	if err != nil {
		return err
	}

	backupInterval, err := time.ParseDuration(vol.Dboxed.BackupInterval)
	if err != nil {
		return err
	}

	logScript := fmt.Sprintf(`#!/bin/sh
exec s6-log n10 s1000000 /var/lib/dboxed/logs/s6/%s
`, name)
	runScript := fmt.Sprintf(`#!/bin/sh
exec dboxed-volume volume serve --dir "%s" --backup-interval="%s" 2>&1
`, workDir, backupInterval.String())

	err = util.AtomicWriteFile(filepath.Join(serviceLogDir, "run"), []byte(logScript), 0755)
	if err != nil {
		return err
	}
	err = util.AtomicWriteFile(filepath.Join(serviceDir, "run"), []byte(runScript), 0755)
	if err != nil {
		return err
	}

	if _, err := os.Lstat(symlinkPath); os.IsNotExist(err) {
		err = os.Symlink(serviceDir, symlinkPath)
		if err != nil {
			return err
		}

		err = vi.s6svscanctl(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (vi volumeInterfaceDboxed) deleteS6Service(ctx context.Context, vol boxspec.BoxVolumeSpec) error {
	name := vi.WorkDirBase(vol)
	serviceDir := filepath.Join("/etc/services.d", name)
	symlinkPath := filepath.Join("/run/service", name)

	if _, err := os.Stat(serviceDir); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	if _, err := os.Lstat(symlinkPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		err = os.RemoveAll(serviceDir)
		if err != nil {
			return err
		}
		return nil
	}

	slog.InfoContext(ctx, "downing s6 service", slog.Any("symlinkPath", symlinkPath))
	err := vi.s6svc(ctx, "-D", "-wd", symlinkPath)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "removing s6 service", slog.Any("symlinkPath", symlinkPath), slog.Any("serviceDir", serviceDir))
	err = os.Remove(symlinkPath)
	if err != nil {
		return err
	}
	err = os.RemoveAll(serviceDir)
	if err != nil {
		return err
	}

	err = vi.s6svscanctl(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (vi volumeInterfaceDboxed) CheckRecreateNeeded(oldVol boxspec.BoxVolumeSpec, newVol boxspec.BoxVolumeSpec) bool {
	return util.MustJson(oldVol) != util.MustJson(newVol)
}

func (vi volumeInterfaceDboxed) runDboxedVolume(ctx context.Context, apiToken string, args []string) error {
	env := os.Environ()
	if apiToken != "" {
		env = append(env, fmt.Sprintf("DBOXED_VOLUME_API_TOKEN=%s", apiToken))
	}

	cmd := exec.CommandContext(ctx, "dboxed-volume", args...)
	cmd.Stdout = vi.rn.DboxedVolumeLog
	cmd.Stderr = vi.rn.DboxedVolumeLog
	cmd.Env = env
	_, _ = fmt.Fprintf(vi.rn.DboxedVolumeLog, "\nrunning: %s\n", cmd.String())
	slog.Info("running dboxed-volume command", slog.Any("args", strings.Join(args, " ")))
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (vi volumeInterfaceDboxed) s6svc(ctx context.Context, args ...string) error {
	slog.InfoContext(ctx, "scanning s6 services")
	cmd := exec.CommandContext(ctx, "/command/s6-svc", args...)
	cmd.Stdout = vi.rn.DboxedVolumeLog
	cmd.Stderr = vi.rn.DboxedVolumeLog
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (vi volumeInterfaceDboxed) s6svscanctl(ctx context.Context) error {
	slog.InfoContext(ctx, "scanning s6 service dir")
	cmd := exec.CommandContext(ctx, "/command/s6-svscanctl", "-h", "/run/service")
	cmd.Stdout = vi.rn.DboxedVolumeLog
	cmd.Stderr = vi.rn.DboxedVolumeLog
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
