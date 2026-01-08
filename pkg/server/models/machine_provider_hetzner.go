package models

import "github.com/dboxed/dboxed/pkg/server/db/dmodel"

type MachineProviderHetzner struct {
	HetznerNetworkName string  `json:"hetznerNetworkName"`
	HetznerNetworkID   *int64  `json:"hetznerNetworkId"`
	HetznerNetworkZone *string `json:"hetznerNetworkZone"`
	HetznerNetworkCidr *string `json:"hetznerNetworkCidr"`
	CloudSubnetCidr    *string `json:"cloudSubnetCidr"`
	RobotSubnetCidr    *string `json:"robotSubnetCidr"`
	RobotVswitchID     *int64  `json:"robotVswitchId"`
}

type UpdateMachineProviderHetzner struct {
	CloudToken    *string `json:"cloudToken,omitempty"`
	RobotUsername *string `json:"robotUsername,omitempty"`
	RobotPassword *string `json:"robotPassword,omitempty"`
}

func MachineProviderHetznerFromDB(v dmodel.MachineProviderHetzner) *MachineProviderHetzner {
	return &MachineProviderHetzner{
		HetznerNetworkName: v.HetznerNetworkName.V,
		HetznerNetworkID:   v.Status.HetznerNetworkID,
		HetznerNetworkZone: v.Status.HetznerNetworkZone,
		HetznerNetworkCidr: v.Status.HetznerNetworkCidr,
		CloudSubnetCidr:    v.Status.CloudSubnetCidr,
		RobotSubnetCidr:    v.Status.RobotSubnetCidr,
		RobotVswitchID:     v.Status.RobotVswitchID,
	}
}
