package run_infra_sandbox

import (
	"encoding/json"
	"github.com/koobox/unboxed/pkg/types"
	"os"
)

func (rn *RunInfraSandbox) getNetbirdStatus() (*types.NetbirdStatus, error) {
	b, err := os.ReadFile(types.NetbirdStatusFile)
	if err != nil {
		return nil, err
	}

	var s types.NetbirdStatus
	err = json.Unmarshal(b, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (rn *RunInfraSandbox) getNetbirdPeerIps() ([]string, error) {
	s, err := rn.getNetbirdStatus()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var ips []string
	for _, p := range s.Peers.Details {
		if p.Status == "Connected" {
			ips = append(ips, p.NetbirdIp)
		}
	}
	return ips, nil
}
