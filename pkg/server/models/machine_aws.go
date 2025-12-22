package models

type CreateMachineAws struct {
	InstanceType   string `json:"instanceType"`
	SubnetId       string `json:"subnetId"`
	RootVolumeSize *int64 `json:"rootVolumeSize,omitempty"`
}

type MachineAws struct {
	InstanceType   string `json:"instanceType"`
	SubnetID       string `json:"subnetID"`
	RootVolumeSize int64  `json:"rootVolumeSize"`

	Status *MachineStatusAws `json:"status,omitempty"`
}

type MachineStatusAws struct {
	InstanceId *string `json:"instanceId,omitempty"`
	PublicIp4  *string `json:"publicIp4,omitempty"`
}
