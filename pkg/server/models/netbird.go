package models

type NetbirdPeerStatus struct {
	NetbirdIp                   string `json:"netbirdIp"`
	PublicKey                   string `json:"publicKey"`
	UsesKernelInterface         bool   `json:"usesKernelInterface"`
	Fqdn                        string `json:"fqdn"`
	QuantumResistance           bool   `json:"quantumResistance"`
	QuantumResistancePermissive bool   `json:"quantumResistancePermissive"`
	ForwardingRules             int    `json:"forwardingRules"`
}
