package types

type InfraConfig struct {
	BoxSpec BoxSpec `json:"boxSpec"`

	BoxName    string `json:"boxName"`
	SandboxDir string `json:"sandboxDir"`

	NetworkConfig NetworkConfig `json:"networkConfig"`
}
