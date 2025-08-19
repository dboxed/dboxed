package types

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/loader"
	ctypes "github.com/compose-spec/compose-go/v2/types"
)

type BoxFile struct {
	Spec BoxSpec `json:"spec"`
}

type BoxSpec struct {
	Uuid string `json:"uuid"`

	DboxedBinaryUrl  string `json:"dboxedBinaryUrl,omitempty"`
	DboxedBinaryHash string `json:"dboxedBinaryHash,omitempty"`

	Dns  *DnsSpec  `json:"dns"`
	Logs *LogsSpec `json:"logs,omitempty"`

	FileBundles     []FileBundle `json:"fileBundles,omitempty"`
	ComposeProjects []string     `json:"composeProjects,omitempty"`
}

type DnsSpec struct {
	Hostname      string `json:"hostname"`
	NetworkDomain string `json:"networkDomain"`
}

type LogsSpec struct {
	Nats *LogsNatsSpec `json:"nats"`
}

type LogsNatsSpec struct {
	MetadataKVStore string `json:"metadataKVStore"`
	LogStream       string `json:"logStream"`
	LogId           string `json:"logId"`
}

type FileBundle struct {
	Name     string            `json:"name"`
	RootUid  uint32            `json:"rootUid"`
	RootGid  uint32            `json:"rootGid"`
	RootMode string            `json:"rootMode"`
	Files    []FileBundleEntry `json:"files"`
}

const AllowedModeMask = os.ModeDir | os.ModeSymlink | os.ModePerm

type FileBundleEntry struct {
	Path string `json:"path"`
	Type string `json:"type,omitempty"` // file, dir, or symlink
	Mode string `json:"mode"`

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

func (s *BoxSpec) LoadComposeProjects() ([]*ctypes.Project, error) {
	var ret []*ctypes.Project
	for _, str := range s.ComposeProjects {
		p, err := s.loadComposeProject(str)
		if err != nil {
			return nil, err
		}
		ret = append(ret, p)
	}
	return ret, nil
}

func (s *BoxSpec) loadComposeProject(str string) (*ctypes.Project, error) {
	tmpFile, err := os.CreateTemp("", "")
	if err != nil {
		return nil, err
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(str))
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
