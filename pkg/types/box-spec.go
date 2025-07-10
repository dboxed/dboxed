package types

import (
	"encoding/base64"
	"fmt"
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

	InfraImage string          `json:"infraImage,omitempty"`
	Containers []ContainerSpec `json:"containers"`

	FileBundles []FileBundle `json:"fileBundles"`
}

type NetbirdSpec struct {
	Version       string `json:"version,omitempty"`
	Image         string `json:"image,omitempty"`
	ManagementUrl string `json:"managementUrl"`
	SetupKey      string `json:"setupKey"`
}

type ContainerSpec struct {
	Name string `json:"name"`

	Image string `json:"image"`

	User             string   `json:"user,omitempty"`
	AdditionalGroups []string `json:"additionalGroups,omitempty"`

	Env        []string `json:"env,omitempty"`
	Entrypoint []string `json:"entrypoint,omitempty"`
	Cmd        []string `json:"cmd,omitempty"`
	WorkingDir string   `json:"workingDir,omitempty"`

	BindMounts   []BindMount   `json:"bindMounts,omitempty"`
	BundleMounts []BundleMount `json:"bundleMounts,omitempty"`

	Privileged  bool `json:"privileged"`
	UseDevTmpFs bool `json:"useDevTmpFs"`

	HostNetwork bool `json:"hostNetwork"`
	HostPid     bool `json:"hostPid"`
	HostCgroups bool `json:"hostCgroups"`
}

type BindMount struct {
	HostPath      string `json:"hostPath"`
	ContainerPath string `json:"containerPath"`
	Shared        bool   `json:"shared"`
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
