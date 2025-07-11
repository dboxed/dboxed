package types

import (
	"encoding/base64"
	"fmt"
	"github.com/compose-spec/compose-go/v2/types"
	"os"
)

type BoxFile struct {
	Spec BoxSpec `json:"spec"`
}

type BoxSpec struct {
	UnboxedBinaryUrl  string `json:"unboxedBinaryUrl,omitempty"`
	UnboxedBinaryHash string `json:"unboxedBinaryHash,omitempty"`

	Netbird NetbirdSpec `json:"netbird"`

	Hostname      string `json:"hostname"`
	NetworkDomain string `json:"networkDomain"`

	InfraImage string `json:"infraImage,omitempty"`

	FileBundles []FileBundle `json:"fileBundles"`

	Compose types.Config `json:"compose"`
}

type NetbirdSpec struct {
	Version       string `json:"version,omitempty"`
	Image         string `json:"image,omitempty"`
	ManagementUrl string `json:"managementUrl"`
	SetupKey      string `json:"setupKey"`
}

type FileBundle struct {
	Name  string            `json:"name"`
	Files []FileBundleEntry `json:"files"`
}

const AllowedModeMask = os.ModeDir | os.ModeSymlink | os.ModePerm

type FileBundleEntry struct {
	Path string `json:"path"`
	Mode uint32 `json:"mode"`

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

type BundleMount struct {
	Name          string `json:"name"`
	ContainerPath string `json:"containerPath"`
}
