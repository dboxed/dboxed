package boxspec

import (
	"net"
)

type NetworkConfig struct {
	SandboxName     string     `json:"sandboxName"`
	VethNetworkCidr *net.IPNet `json:"vethNetworkCidr"`
}

type NetworkHost struct {
	Name string `json:"name"`
	IP4  string `json:"ip4"`
}
