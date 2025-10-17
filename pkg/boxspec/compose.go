package boxspec

import (
	"context"
	"fmt"
	"os"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/loader"
	ctypes "github.com/compose-spec/compose-go/v2/types"
)

type GetMountFunc func(volumeUuid string) string

func (s *BoxSpec) LoadComposeProjects(ctx context.Context, getMount GetMountFunc) ([]*ctypes.Project, error) {
	var ret []*ctypes.Project
	for name, str := range s.ComposeProjects {
		p, err := s.loadAndSetupComposeProject(ctx, name, str, getMount)
		if err != nil {
			return nil, err
		}
		ret = append(ret, p)
	}
	return ret, nil
}

func (s *BoxSpec) ValidateComposeProjects(ctx context.Context) error {
	for _, str := range s.ComposeProjects {
		err := s.ValidateComposeProject(ctx, str)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *BoxSpec) loadAndSetupComposeProject(ctx context.Context, name string, composeStr string, getMount GetMountFunc) (*ctypes.Project, error) {
	cp, err := s.loadComposeProject(ctx, name, composeStr, false)
	if err != nil {
		return nil, err
	}

	err = s.setupVolumes(cp, getMount)
	if err != nil {
		return nil, err
	}

	return cp, nil
}

func (s *BoxSpec) ValidateComposeProject(ctx context.Context, composeStr string) error {
	getMount := func(volumeUuid string) string {
		return "/dummy"
	}
	cp, err := s.loadAndSetupComposeProject(ctx, "dummy", composeStr, getMount)
	if err != nil {
		return err
	}

	str2, err := cp.MarshalYAML()
	if err != nil {
		return err
	}
	_, err = s.loadComposeProject(ctx, "dummy", string(str2), true)
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

	x.Name = name

	return x, nil
}

func (s *BoxSpec) setupVolumes(compose *ctypes.Project, getMount GetMountFunc) error {
	volumesByName := map[string]*DboxedVolume{}
	for _, vol := range s.Volumes {
		volumesByName[vol.Name] = &vol
	}

	for _, service := range compose.Services {
		for i, _ := range service.Volumes {
			volume := &service.Volumes[i]
			if volume.Type == "dboxed" {
				vol, ok := volumesByName[volume.Source]
				if !ok {
					return fmt.Errorf("volume with name %s not found", volume.Source)
				}

				volume.Type = ctypes.VolumeTypeBind
				if getMount != nil {
					volume.Source = getMount(vol.Uuid)
				} else {
					volume.Source = "/dummy"
				}
			}
		}
	}
	return nil
}
