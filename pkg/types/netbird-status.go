package types

import "time"

type NetbirdStatus struct {
	Name      string `json:"name"`
	NetbirdIp string `json:"netbirdIp"`

	Peers NetbirdPeers `json:"peers"`
}

type NetbirdPeers struct {
	Total     int                 `json:"total"`
	Connected int                 `json:"connected"`
	Details   []NetbirdPeerStatus `json:"details"`
}

type NetbirdPeerStatus struct {
	Name                   string    `json:"name"`
	FQDN                   string    `json:"fqdn"`
	NetbirdIp              string    `json:"netbirdIp"`
	PublicKey              string    `json:"publicKey"`
	Status                 string    `json:"status"`
	LastStatusUpdate       time.Time `json:"lastStatusUpdate"`
	ConnectionType         string    `json:"connectionType"`
	RelayAddress           string    `json:"relayAddress"`
	LastWireguardHandshake time.Time `json:"lastWireguardHandshake"`
	TransferReceived       int64     `json:"transferReceived"`
	TransferSent           int64     `json:"transferSent"`
	Latency                int64     `json:"latency"`
}
