//go:build linux

package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dboxed/dboxed/pkg/util"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/opencontainers/runc/libcontainer"
)

type S6Service struct {
	ServiceName   string
	RunContent    string
	RunLogContent string

	s6Helper S6Helper
}

func (s *S6Service) Install(ctx context.Context) error {
	err := s.s6Helper.detectS6Dirs()
	if err != nil {
		return err
	}

	serviceDir := filepath.Join(s.s6Helper.Rootfs, s.s6Helper.ServicesDir, s.ServiceName)
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

	symlinkPath := filepath.Join(s.s6Helper.ServiceLinksDir, s.ServiceName)
	err = os.Symlink(serviceDir, symlinkPath)
	if err != nil {
		return err
	}

	// initially disable it
	err = os.WriteFile(filepath.Join(serviceDir, "down"), nil, 0644)
	if err != nil {
		return err
	}

	err = s.s6Helper.S6svscanctl(ctx, s.s6Helper.ServiceLinksDir)
	if err != nil {
		return err
	}

	return nil
}

func (s *S6Service) Uninstall(ctx context.Context) error {
	err := s.s6Helper.detectS6Dirs()
	if err != nil {
		return err
	}

	err = s.Stop(ctx)
	if err != nil {
		return err
	}

	serviceDir := filepath.Join(s.s6Helper.Rootfs, s.s6Helper.ServicesDir, s.ServiceName)
	symlinkPath := filepath.Join(s.s6Helper.Rootfs, s.s6Helper.ServiceLinksDir, s.ServiceName)

	err = os.Remove(symlinkPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	err = os.RemoveAll(serviceDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}

func (s *S6Service) Enable(ctx context.Context) error {
	err := s.s6Helper.detectS6Dirs()
	if err != nil {
		return err
	}

	serviceDir := filepath.Join(s.s6Helper.Rootfs, s.s6Helper.ServicesDir, s.ServiceName)
	_ = os.Remove(filepath.Join(serviceDir, "down"))

	return nil
}

func (s *S6Service) Start(ctx context.Context) error {
	err := s.s6Helper.S6SvcUp(ctx, s.ServiceName)
	if err != nil {
		return err
	}
	return nil
}

func (s *S6Service) Stop(ctx context.Context) error {
	err := s.s6Helper.S6SvcDown(ctx, s.ServiceName)
	if err != nil {
		return err
	}
	return nil
}

type S6Helper struct {
	Container   *libcontainer.Container
	ImageConfig *v1.ImageConfig

	ServicesDir     string
	ServiceLinksDir string

	Rootfs string

	detectOnce sync.Once
	detectErr  error
}

func (s *S6Helper) detectS6Dirs() error {
	s.detectOnce.Do(func() {
		s.detectErr = s.doDetectS6Dirs()
	})
	return s.detectErr
}

func (s *S6Helper) doDetectS6Dirs() error {
	if s.Container != nil {
		s.Rootfs = s.Container.Config().Rootfs
	}

	checkDir := func(dir string) (bool, error) {
		dir = filepath.Join(s.Rootfs, dir)

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

	if ok, err := checkDir("/etc/services.d"); ok {
		s.ServicesDir = "/etc/services.d"
	} else if err != nil {
		return err
	} else {
		return fmt.Errorf("failed to determine s6 services dir")
	}

	if ok, err := checkDir("/run/service"); ok {
		s.ServiceLinksDir = "/run/service"
	} else if err != nil {
		return err
	} else {
		return fmt.Errorf("failed to determine s6 services links dir")
	}

	return nil
}

func (s *S6Helper) S6svc(ctx context.Context, serviceName string, args ...string) error {
	err := s.detectS6Dirs()
	if err != nil {
		return err
	}

	serviceDir := filepath.Join(s.ServiceLinksDir, serviceName)

	var args2 []string
	args2 = append(args2, args...)
	args2 = append(args2, serviceDir)

	cmd := util.CommandHelper{
		ContainerHolder: util.ContainerHolder{
			Container:   s.Container,
			ImageConfig: s.ImageConfig,
		},
		Command: "s6-svc",
		Args:    args2,
		LogCmd:  true,
	}
	err = cmd.Run(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *S6Helper) S6svscanctl(ctx context.Context, servicesDir string) error {
	slog.InfoContext(ctx, "scanning s6 services dir")
	cmd := util.CommandHelper{
		ContainerHolder: util.ContainerHolder{
			Container:   s.Container,
			ImageConfig: s.ImageConfig,
		},
		Command: "s6-svscanctl",
		Args:    []string{"-h", servicesDir},
		LogCmd:  true,
	}
	err := cmd.Run(ctx)
	if err != nil {
		return err
	}
	// A race comdition in s6 causes follow up errors when running s6-svc too early after this
	// see https://github.com/just-containers/s6-overlay/issues/460
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (s *S6Helper) S6SvcUp(ctx context.Context, serviceName string) error {
	return s.S6svc(ctx, serviceName, "-u", "-wu")
}

func (s *S6Helper) S6SvcDown(ctx context.Context, serviceName string) error {
	if !s.S6SvcExists(serviceName) {
		return nil
	}

	return s.S6svc(ctx, serviceName, "-d", "-wd")
}

func (s *S6Helper) S6SvcRestart(ctx context.Context, serviceName string) error {
	return s.S6svc(ctx, serviceName, "-r", "-wr")
}

func (s *S6Helper) S6SvcExists(serviceName string) bool {
	err := s.detectS6Dirs()
	if err != nil {
		return false
	}

	serviceDir := filepath.Join(s.Rootfs, s.ServiceLinksDir, serviceName)
	slog.Info("check", slog.Any("serviceDir", serviceDir))
	_, err = os.Lstat(serviceDir)
	if err != nil {
		return false
	}
	return true
}
