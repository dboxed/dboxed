package volume_serve

import (
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
	"sigs.k8s.io/yaml"
)

type VolumeState struct {
	ClientAuth *baseclient.ClientAuth `json:"clientAuth,omitempty"`

	MountName string `json:"mountName"`

	Volume *models.Volume `json:"volume"`

	RestoreDone bool `json:"restoreDone"`
}

func (vs *VolumeServe) loadVolumeState() (*VolumeState, error) {
	s, err := LoadVolumeState(vs.opts.Dir)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func LoadVolumeState(dir string) (*VolumeState, error) {
	s, err := util.UnmarshalYamlFile[VolumeState](filepath.Join(dir, "volume-state.yaml"))
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (vs *VolumeServe) saveVolumeState(s VolumeState) error {
	b, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return util.AtomicWriteFile(filepath.Join(vs.opts.Dir, "volume-state.yaml"), b, 0600)
}

func ListVolumeState(baseDir string) ([]*VolumeState, error) {
	des, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var ret []*VolumeState
	for _, de := range des {
		if !de.IsDir() {
			continue
		}
		dir := filepath.Join(baseDir, de.Name())
		info, err := LoadVolumeState(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		ret = append(ret, info)
	}

	return ret, nil
}
