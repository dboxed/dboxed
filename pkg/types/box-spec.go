package types

type BoxFile struct {
	Spec BoxSpec `json:"spec"`
}

type BoxSpec struct {
	UnboxedBinaryUrl  string `json:"unboxedBinaryUrl"`
	UnboxedBinaryHash string `json:"unboxedBinaryHash"`

	Netbird NetbirdSpec `json:"netbird"`

	Hostname      string `json:"hostname"`
	NetworkDomain string `json:"networkDomain"`

	InfraImage string          `json:"infraImage"`
	Containers []ContainerSpec `json:"containers"`
}

type NetbirdSpec struct {
	ManagementUrl string `json:"managementUrl"`
	SetupKey      string `json:"setupKey"`
}

type ContainerSpec struct {
	Name string `json:"name"`

	Image string `json:"image"`

	User       string   `json:"user"`
	Env        []string `json:"env"`
	Entrypoint []string `json:"entrypoint"`
	Cmd        []string `json:"cmd"`
	WorkingDir string   `json:"workingDir"`

	Privileged  bool `json:"privileged"`
	UseDevTmpFs bool `json:"useDevTmpFs"`
}
