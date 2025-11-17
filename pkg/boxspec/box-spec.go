package boxspec

import "time"

type BoxSpec struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	DesiredState string `json:"desiredState"`

	ReconcileRequestedAt *time.Time `json:"reconcileRequestedAt,omitempty"`

	Network *BoxNetwork    `json:"network,omitempty"`
	Volumes []DboxedVolume `json:"volumes,omitempty"`

	ComposeProjects map[string]string `json:"composeProjects,omitempty"`
}

type BoxNetwork struct {
	ID   *string `json:"ID"`
	Name *string `json:"name,omitempty"`

	Netbird *BoxNetworkNetbird `json:"netbird,omitempty"`

	PortForwards []PortForward `json:"portForwards,omitempty"`
	NetworkHosts []NetworkHost `json:"networkHosts,omitempty"`
}

type BoxNetworkNetbird struct {
	ManagementUrl string `json:"managementUrl"`
	SetupKey      string `json:"setupKey"`
	Hostname      string `json:"hostname"`
}
