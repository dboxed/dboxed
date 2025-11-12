package boxspec

type PortForward struct {
	Protocol      string `json:"protocol"`
	HostFirstPort int    `json:"hostFirstPort"`
	HostLastPort  int    `json:"hostLastPort"`

	SandboxPort int `json:"sandboxPort"`
}
