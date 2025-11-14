package boxspec

type NetworkConfig struct {
	SandboxName     string `json:"sandboxName"`
	VethNetworkCidr string `json:"vethNetworkCidr"`
}

type NetworkHost struct {
	Name string `json:"name"`
	IP4  string `json:"ip4"`
}
