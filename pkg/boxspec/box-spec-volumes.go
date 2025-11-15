package boxspec

import (
	"os"
)

type DboxedVolume struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	RootUid  uint32 `json:"rootUid"`
	RootGid  uint32 `json:"rootGid"`
	RootMode string `json:"rootMode"`

	BackupInterval string `json:"backupInterval"`
}

const AllowedModeMask = os.ModePerm

func (s *BoxSpec) GetVolumeByName(name string) *DboxedVolume {
	for _, v := range s.Volumes {
		if v.Name == name {
			return &v
		}
	}
	return nil
}
