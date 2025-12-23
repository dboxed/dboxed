package models

type CreateMachineHetzner struct {
	ServerType     string `json:"serverType"`
	ServerLocation string `json:"serverLocation"`
}

type MachineHetzner struct {
	ServerType     string `json:"serverType"`
	ServerLocation string `json:"serverLocation"`

	Status *MachineStatusHetzner `json:"status,omitempty"`
}

type MachineStatusHetzner struct {
	ServerId  *int64  `json:"serverId,omitempty"`
	PublicIp4 *string `json:"publicIp4,omitempty"`
}
