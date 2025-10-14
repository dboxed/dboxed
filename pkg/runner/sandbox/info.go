package sandbox

import (
	"net/netip"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
)

type SandboxInfo struct {
	SandboxName     string
	Box             *models.Box       `json:"box"`
	Workspace       *models.Workspace `json:"workspace"`
	VethNetworkCidr string            `json:"vethNetworkCidr"`
}

func ListSandboxes(sandboxBaseDir string) ([]SandboxInfo, error) {
	des, err := os.ReadDir(sandboxBaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var ret []SandboxInfo
	for _, de := range des {
		if !de.IsDir() {
			continue
		}

		sandboxDir := filepath.Join(sandboxBaseDir, de.Name())
		si, err := ReadSandboxInfo(sandboxDir)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
			continue
		}
		ret = append(ret, *si)
	}
	return ret, nil
}

func ReadSandboxInfo(sandboxDir string) (*SandboxInfo, error) {
	si, err := util.UnmarshalYamlFile[SandboxInfo](filepath.Join(sandboxDir, consts.SandboxInfoFile))
	if err != nil {
		return nil, err
	}
	return si, nil
}

func WriteSandboxInfo(sandboxDir string, si *SandboxInfo) error {
	err := util.AtomicWriteFileYaml(filepath.Join(sandboxDir, consts.SandboxInfoFile), si, 0600)
	if err != nil {
		return err
	}
	return nil
}

func ReadVethCidr(sandboxDir string) (*netip.Prefix, error) {
	pth := filepath.Join(sandboxDir, consts.VethIPStoreFile)
	ipB, err := os.ReadFile(pth)
	if err != nil {
		return nil, err
	}
	p, err := netip.ParsePrefix(string(ipB))
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func WriteVethCidr(sandboxDir string, p *netip.Prefix) error {
	pth := filepath.Join(sandboxDir, consts.VethIPStoreFile)
	return util.AtomicWriteFile(pth, []byte(p.String()), 0644)
}
