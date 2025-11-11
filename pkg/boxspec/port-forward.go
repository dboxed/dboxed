package boxspec

type PortForward struct {
	IP string `json:"ip"`

	Protocol      string `json:"protocol"`
	HostFirstPort int    `json:"hostFirstPort"`
	HostLastPort  int    `json:"hostLastPort"`

	SandboxPort int `json:"sandboxPort"`
}
