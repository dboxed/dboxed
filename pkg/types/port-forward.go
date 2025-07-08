package types

type PortForward struct {
	IP string `json:"ip"`

	FromPort int    `json:"fromPort"`
	ToPort   int    `json:"toPort"`
	Protocol string `json:"protocol"`

	DestinationPort int `json:"destinationPort"`
}
