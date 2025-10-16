package boxspec

import (
	"context"
	"os"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/loader"
	ctypes "github.com/compose-spec/compose-go/v2/types"
)

type BoxSpec struct {
	Uuid string `json:"uuid"`

	Volumes []DboxedVolume `json:"volumes,omitempty"`

	ComposeProjects []string `json:"composeProjects,omitempty"`
}

func (s *BoxSpec) LoadComposeProjects() ([]*ctypes.Project, error) {
	var ret []*ctypes.Project
	for _, str := range s.ComposeProjects {
		p, err := s.loadComposeProject(str)
		if err != nil {
			return nil, err
		}
		ret = append(ret, p)
	}
	return ret, nil
}

func (s *BoxSpec) loadComposeProject(str string) (*ctypes.Project, error) {
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

	options, err := cli.NewProjectOptions(
		[]string{tmpFile.Name()},
		// we need to skip validation as we're using "bundle" volumes, which are not valid as by the spec
		cli.WithLoadOptions(loader.WithSkipValidation),
		cli.WithNormalization(false),
		cli.WithInterpolation(false),
		cli.WithConsistency(false),
		cli.WithResolvedPaths(false),
		cli.WithoutEnvironmentResolution,
	)
	if err != nil {
		return nil, err
	}
	x, err := options.LoadProject(context.Background())
	if err != nil {
		return nil, err
	}
	return x, nil
}
