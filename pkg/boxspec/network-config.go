package boxspec

import (
	"net"
)

type NetworkConfig struct {
	SandboxName     string     `json:"sandboxName"`
	VethNetworkCidr *net.IPNet `json:"vethNetworkCidr"`
	DnsProxyIP      string     `json:"dnsProxyIP"`
}
