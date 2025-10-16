package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dboxed/dboxed/pkg/util"
)

type S6Service struct {
	ServiceName   string
	RunContent    string
	RunLogContent string
}

func (s *S6Service) Install(ctx context.Context) error {
	dirs, err := DetectS6Dirs()
	if err != nil {
		return err
	}

	serviceDir := filepath.Join(dirs.ServicesDir, s.ServiceName)
	symlinkPath := filepath.Join(dirs.ServiceLinksDir, s.ServiceName)
	serviceLogDir := filepath.Join(serviceDir, "log")

	err = os.MkdirAll(serviceDir, 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(serviceLogDir, 0755)
	if err != nil {
		return err
	}

	err = util.AtomicWriteFile(filepath.Join(serviceDir, "run"), []byte(s.RunContent), 0755)
	if err != nil {
		return err
	}
	if s.RunLogContent != "" {
		err = util.AtomicWriteFile(filepath.Join(serviceLogDir, "run"), []byte(s.RunLogContent), 0755)
		if err != nil {
			return err
		}
	}

	err = os.Symlink(serviceDir, symlinkPath)
	if err != nil {
		return err
	}

	// initially disable it
	err = os.WriteFile(filepath.Join(serviceDir, "down"), nil, 0644)
	if err != nil {
		return err
	}

	err = s6svscanctl(ctx, dirs.ServiceLinksDir)
	if err != nil {
		return err
	}

	return nil
}

func (s *S6Service) Uninstall(ctx context.Context) error {
	dirs, err := DetectS6Dirs()
	if err != nil {
		return err
	}

	err = s.Stop(ctx)
	if err != nil {
		return err
	}

	serviceDir := filepath.Join(dirs.ServicesDir, s.ServiceName)
	symlinkPath := filepath.Join(dirs.ServiceLinksDir, s.ServiceName)

	err = os.Remove(symlinkPath)
	if err != nil {
		return err
	}
	err = os.RemoveAll(serviceDir)
	if err != nil {
		return err
	}

	return nil
}

func (s *S6Service) Enable(ctx context.Context) error {
	dirs, err := DetectS6Dirs()
	if err != nil {
		return err
	}

	serviceDir := filepath.Join(dirs.ServicesDir, s.ServiceName)
	_ = os.Remove(filepath.Join(serviceDir, "down"))

	return nil
}

func (s *S6Service) Start(ctx context.Context) error {
	dirs, err := DetectS6Dirs()
	if err != nil {
		return err
	}

	symlinkPath := filepath.Join(dirs.ServiceLinksDir, s.ServiceName)

	err = s6svc(ctx, "-u", "-wu", symlinkPath)
	if err != nil {
		return err
	}
	return nil
}

func (s *S6Service) Stop(ctx context.Context) error {
	dirs, err := DetectS6Dirs()
	if err != nil {
		return err
	}

	symlinkPath := filepath.Join(dirs.ServiceLinksDir, s.ServiceName)

	err = s6svc(ctx, "-d", "-wd", symlinkPath)
	if err != nil {
		return err
	}
	return nil
}

type S6Dirs struct {
	ServicesDir     string
	ServiceLinksDir string
}

func DetectS6Dirs() (*S6Dirs, error) {
	checkDir := func(dir string) (bool, error) {
		st, err := os.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, err
		}
		if !st.IsDir() {
			return false, fmt.Errorf("%s is not a directory", dir)
		}
		return true, nil
	}

	var ret S6Dirs
	if ok, err := checkDir("/etc/services.d"); ok {
		ret.ServicesDir = "/etc/services.d"
	} else if err != nil {
		return nil, err
	} else {
		return nil, fmt.Errorf("failed to determine s6 services dir")
	}

	if ok, err := checkDir("/run/service"); ok {
		ret.ServiceLinksDir = "/run/service"
	} else if err != nil {
		return nil, err
	} else {
		return nil, fmt.Errorf("failed to determine s6 services links dir")
	}

	return &ret, nil
}

func s6svc(ctx context.Context, args ...string) error {
	slog.InfoContext(ctx, fmt.Sprintf("running s6-svc %s", strings.Join(args, " ")))
	cmd := exec.CommandContext(ctx, "s6-svc", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func s6svscanctl(ctx context.Context, servicesDir string) error {
	slog.InfoContext(ctx, "scanning s6 services dir")
	cmd := exec.CommandContext(ctx, "s6-svscanctl", "-h", servicesDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	// A race comdition in s6 causes follow up errors when running s6-svc too early after this
	// see https://github.com/just-containers/s6-overlay/issues/460
	time.Sleep(100 * time.Millisecond)
	return nil
}
