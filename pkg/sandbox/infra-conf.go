package sandbox

import (
	"encoding/json"
	"fmt"
	"github.com/koobox/unboxed/pkg/types"
	"github.com/koobox/unboxed/pkg/util"
	"os"
	"path/filepath"
)

func (rn *Sandbox) writeInfraConf() error {
	infraConf := types.InfraConfig{
		BoxSpec:              *rn.BoxSpec,
		BoxName:              rn.SandboxName,
		SandboxDir:           rn.SandboxDir,
		NetworkNamespaceName: rn.network.NetworkNamespaceName,
		DnsProxyIP:           rn.network.PeerAddr.IP.String(),
	}

	b, err := json.Marshal(infraConf)
	if err != nil {
		return err
	}

	err = util.AtomicWriteFile(filepath.Join(rn.getSharedDirOnHost(), filepath.Base(types.InfraConfFile)), b, 0600)
	if err != nil {
		return fmt.Errorf("failed to write infra conf into shared dir: %w", err)
	}

	return nil
}

func ReadInfraConf(path string) (*types.InfraConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read infra conf: %w", err)
	}

	var c types.InfraConfig
	err = json.Unmarshal(b, &c)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal infra conf: %w", err)
	}
	return &c, nil
}
