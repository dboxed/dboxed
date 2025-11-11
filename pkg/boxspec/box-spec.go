package boxspec

type BoxSpec struct {
	ID           string `json:"id"`
	DesiredState string `json:"desiredState"`

	Network *BoxNetwork    `json:"network,omitempty"`
	Volumes []DboxedVolume `json:"volumes,omitempty"`

	ComposeProjects map[string]string `json:"composeProjects,omitempty"`
}

type BoxNetwork struct {
	Netbird *BoxNetworkNetbird `json:"netbird,omitempty"`
}

type BoxNetworkNetbird struct {
	ManagementUrl string `json:"managementUrl"`
	SetupKey      string `json:"setupKey"`
	Hostname      string `json:"hostname"`
}
