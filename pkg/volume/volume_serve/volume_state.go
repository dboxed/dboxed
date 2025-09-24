package volume_serve

import (
	"encoding/json"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/util"
)

type VolumeState struct {
	ClientAuth *baseclient.ClientAuth `json:"clientAuth,omitempty"`

	VolumeId   int64   `json:"volumeId"`
	VolumeUuid string  `json:"volumeUuid"`
	LockId     *string `json:"lockId,omitempty"`
}

func (vs *VolumeServe) loadVolumeState() (*VolumeState, error) {
	s, err := LoadVolumeState(vs.opts.Dir)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func LoadVolumeState(dir string) (*VolumeState, error) {
	s, err := util.UnmarshalJsonFile[VolumeState](filepath.Join(dir, "volume-state.json"))
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (vs *VolumeServe) saveVolumeState(s VolumeState) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return util.AtomicWriteFile(filepath.Join(vs.opts.Dir, "volume-state.json"), b, 0600)
}
