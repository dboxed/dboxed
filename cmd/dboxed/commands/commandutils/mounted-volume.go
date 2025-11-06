package commandutils

import (
	"os"

	"github.com/dboxed/dboxed/pkg/volume/volume_serve"
)

func GetMountedVolume(baseDir string, volume string) (*volume_serve.VolumeState, error) {
	mountedVolumes, err := volume_serve.ListVolumeState(baseDir)
	if err != nil {
		return nil, err
	}

	for _, mv := range mountedVolumes {
		if mv.Volume == nil || mv.Volume.Name == volume || mv.Volume.ID == volume {
			return mv, nil
		}
	}
	return nil, os.ErrNotExist
}
