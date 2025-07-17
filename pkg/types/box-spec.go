package types

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/loader"
	ctypes "github.com/compose-spec/compose-go/v2/types"
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

	Compose string `json:"compose"`
}

type NetbirdSpec struct {
	Version       string `json:"version,omitempty"`
	Image         string `json:"image,omitempty"`
	ManagementUrl string `json:"managementUrl"`
	SetupKey      string `json:"setupKey"`
}

type FileBundle struct {
	Name     string            `json:"name"`
	RootUid  uint32            `json:"rootUid"`
	RootGid  uint32            `json:"rootGid"`
	RootMode uint32            `json:"rootMode"`
	Files    []FileBundleEntry `json:"files"`
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

func (s *BoxSpec) LoadComposeProject() (*ctypes.Project, error) {
	tmpFile, err := os.CreateTemp("", "")
	if err != nil {
		return nil, err
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(s.Compose))
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
