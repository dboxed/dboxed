package types

import (
	"net"
)

type NetworkConfig struct {
	SandboxName     string
	VethNetworkCidr *net.IPNet
	DnsProxyIP      string
}
