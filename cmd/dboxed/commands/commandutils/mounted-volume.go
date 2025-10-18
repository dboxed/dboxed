package commandutils

import (
	"os"
	"strconv"

	"github.com/dboxed/dboxed/pkg/volume/volume_serve"
)

func GetMountedVolume(baseDir string, volume string) (*volume_serve.VolumeState, error) {
	mountedVolumes, err := volume_serve.ListVolumeState(baseDir)
	if err != nil {
		return nil, err
	}

	volumeId, err := strconv.ParseInt(volume, 10, 64)
	if err != nil {
		volumeId = -1
	}

	for _, mv := range mountedVolumes {
		if mv.Volume == nil || mv.Volume.Name == volume || mv.Volume.ID == volumeId || mv.Volume.Uuid == volume {
			return mv, nil
		}
	}
	return nil, os.ErrNotExist
}
