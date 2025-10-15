package boxspec

import (
	"os"
)

type DboxedVolume struct {
	Uuid string `json:"uuid"`
	Name string `json:"name"`
	Id   int64  `json:"id"`

	RootUid  uint32 `json:"rootUid"`
	RootGid  uint32 `json:"rootGid"`
	RootMode string `json:"rootMode"`

	BackupInterval string `json:"backupInterval"`
}

const AllowedModeMask = os.ModePerm
