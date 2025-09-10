package types

import (
	"encoding/base64"
	"fmt"
	"os"
)

type BoxVolumeSpec struct {
	Name string `json:"name"`

	RootUid  uint32 `json:"rootUid"`
	RootGid  uint32 `json:"rootGid"`
	RootMode string `json:"rootMode"`

	FileBundle *FileBundle   `json:"fileBundle,omitempty"`
	Dboxed     *DboxedVolume `json:"dboxed,omitempty"`
}

type FileBundle struct {
	Files []FileBundleEntry `json:"files"`
}

type DboxedVolume struct {
	ApiUrl string `json:"apiUrl"`
	Token  string `json:"token"`

	RepositoryId int64 `json:"repoId"`
	VolumeId     int64 `json:"volumeId"`

	BackupInterval string `json:"backupInterval"`
}

const AllowedModeMask = os.ModePerm

type FileBundleEntryType string

const (
	FileBundleEntryFile    FileBundleEntryType = "file"
	FileBundleEntryDir     FileBundleEntryType = "dir"
	FileBundleEntrySymlink FileBundleEntryType = "symlink"
)

type FileBundleEntry struct {
	Path string              `json:"path"`
	Type FileBundleEntryType `json:"type,omitempty"` // file, dir, or symlink
	Mode string              `json:"mode"`

	Uid int `json:"uid"`
	Gid int `json:"gid"`

	// Data must be base64 encoded
	Data string `json:"data,omitempty"`

	// StringData is an alternative to Data
	StringData string `json:"stringData,omitempty"`
}

func (e *FileBundleEntry) GetDecodedData() ([]byte, error) {
	if e.Data == "" && e.StringData == "" {
		return nil, nil
	}
	if e.Data != "" && e.StringData != "" {
		return nil, fmt.Errorf("both data and stringData are set")
	} else if e.Data != "" {
		x, err := base64.StdEncoding.DecodeString(e.Data)
		if err != nil {
			return nil, err
		}
		return x, nil
	} else {
		return []byte(e.StringData), nil
	}
}
