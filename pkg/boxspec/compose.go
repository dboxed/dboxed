package boxspec

import (
	"context"
	"fmt"
	"os"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/loader"
	ctypes "github.com/compose-spec/compose-go/v2/types"
)

type UpdateServiceVolumeFunc func(volume *ctypes.ServiceVolumeConfig) error

func (s *BoxSpec) LoadComposeProjects(ctx context.Context, updateServiceVolume UpdateServiceVolumeFunc) (map[string]*ctypes.Project, error) {
	ret := map[string]*ctypes.Project{}
	for name, str := range s.ComposeProjects {
		p, err := s.loadComposeProject(ctx, name, str, false)
		if err != nil {
			return nil, err
		}
		if updateServiceVolume != nil {
			err = s.setupDboxedVolumesForProject(p, updateServiceVolume)
			if err != nil {
				return nil, err
			}
		}
		ret[name] = p
	}
	return ret, nil
}

func (s *BoxSpec) ValidateComposeProjects(ctx context.Context) error {
	for name, str := range s.ComposeProjects {
		err := s.ValidateComposeProject(ctx, name, str)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *BoxSpec) loadAndSetupComposeProject(ctx context.Context, name string, composeStr string, updateServiceVolume UpdateServiceVolumeFunc) (*ctypes.Project, error) {
	cp, err := s.loadComposeProject(ctx, name, composeStr, false)
	if err != nil {
		return nil, err
	}

	err = s.setupDboxedVolumesForProject(cp, updateServiceVolume)
	if err != nil {
		return nil, err
	}

	return cp, nil
}

func (s *BoxSpec) ValidateComposeProject(ctx context.Context, name string, composeStr string) error {
	updateServiceVolume := func(volume *ctypes.ServiceVolumeConfig) error {
		volume.Type = ctypes.VolumeTypeBind
		volume.Source = "/dummy"
		return nil
	}
	cp, err := s.loadComposeProject(ctx, name, composeStr, false)
	if err != nil {
		return err
	}
	err = s.setupDboxedVolumesForProject(cp, updateServiceVolume)
	if err != nil {
		return err
	}

	str2, err := cp.MarshalYAML()
	if err != nil {
		return err
	}
	_, err = s.loadComposeProject(ctx, name, string(str2), true)
	if err != nil {
		return err
	}
	return nil
}

func (s *BoxSpec) loadComposeProject(ctx context.Context, name string, str string, validate bool) (*ctypes.Project, error) {
	tmpFile, err := os.CreateTemp("", "")
	if err != nil {
		return nil, err
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(str))
	if err != nil {
		return nil, err
	}
	err = tmpFile.Close()
	if err != nil {
		return nil, err
	}

	var opts []cli.ProjectOptionsFn
	if !validate {
		// we need to skip validation as we're using "bundle" volumes, which are not valid as by the spec
		opts = append(opts,
			cli.WithLoadOptions(loader.WithSkipValidation),
			cli.WithNormalization(false),
			cli.WithInterpolation(false),
			cli.WithConsistency(false),
			cli.WithResolvedPaths(false),
			cli.WithoutEnvironmentResolution,
		)
	}

	options, err := cli.NewProjectOptions([]string{tmpFile.Name()}, opts...)
	if err != nil {
		return nil, err
	}
	x, err := options.LoadProject(ctx)
	if err != nil {
		return nil, err
	}

	if x.Name != "" && x.Name != name {
		return nil, fmt.Errorf("name in compose project file does not match expected name")
	}
	x.Name = name

	return x, nil
}

func (s *BoxSpec) setupDboxedVolumesForProject(compose *ctypes.Project, updateServiceVolume UpdateServiceVolumeFunc) error {
	for _, service := range compose.Services {
		for i, _ := range service.Volumes {
			volume := &service.Volumes[i]
			err := updateServiceVolume(volume)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
