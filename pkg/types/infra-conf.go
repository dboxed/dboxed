package types

type InfraConfig struct {
	BoxSpec BoxSpec `json:"boxSpec"`

	BoxName    string `json:"boxName"`
	SandboxDir string `json:"sandboxDir"`

	NetworkNamespaceName string `json:"networkNamespaceName"`

	DnsProxyIP string `json:"dnsProxyIP"`
}
