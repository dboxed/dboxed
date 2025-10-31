package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/global"
)

type MachineProvider struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Workspace int64     `json:"workspace"`

	Status        string `json:"status"`
	StatusDetails string `json:"statusDetails"`

	Type string `json:"type"`
	Name string `json:"name"`

	SshKeyFingerprint *string `json:"sshKeyFingerprint"`

	Aws     *MachineProviderAws     `json:"aws,omitempty"`
	Hetzner *MachineProviderHetzner `json:"hetzner,omitempty"`
}

type MachineProviderAws struct {
	Region          string  `json:"region"`
	VpcID           *string `json:"vpcId"`
	VpcName         *string `json:"vpcName"`
	VpcCidr         *string `json:"vpcCidr"`
	SecurityGroupID *string `json:"securityGroupId"`

	Subnets []MachineProviderAwsSubnet `json:"subnets"`
}

type MachineProviderAwsSubnet struct {
	MachineProvider  int64   `json:"machineProvider"`
	SubnetID         string  `json:"subnetId"`
	SubnetName       *string `json:"subnetName"`
	AvailabilityZone string  `json:"availabilityZone"`
	Cidr             string  `json:"cidr"`
}

type MachineProviderHetzner struct {
	HetznerNetworkName string  `json:"hetznerNetworkName"`
	HetznerNetworkID   *int64  `json:"hetznerNetworkId"`
	HetznerNetworkZone *string `json:"hetznerNetworkZone"`
	HetznerNetworkCidr *string `json:"hetznerNetworkCidr"`
	CloudSubnetCidr    *string `json:"cloudSubnetCidr"`
	RobotSubnetCidr    *string `json:"robotSubnetCidr"`
	RobotVswitchID     *int64  `json:"robotVswitchId"`
}

type CreateMachineProvider struct {
	Type global.MachineProviderType `json:"type"`
	Name string                     `json:"name"`

	SshKeyPublic *string `json:"sshKeyPublic,omitempty"`

	Aws     *CreateMachineProviderAws     `json:"aws,omitempty"`
	Hetzner *CreateMachineProviderHetzner `json:"hetzner,omitempty"`
}

type CreateMachineProviderAws struct {
	Region string `json:"region"`
	VpcId  string `json:"vpcId"`

	AwsAccessKeyId     string `json:"awsAccessKeyId"`
	AwsSecretAccessKey string `json:"awsSecretAccessKey"`
}

type CreateMachineProviderHetzner struct {
	CloudToken    string  `json:"cloudToken"`
	RobotUsername *string `json:"robotUsername,omitempty"`
	RobotPassword *string `json:"robotPassword,omitempty"`

	HetznerNetworkName string `json:"hetznerNetworkName"`
}

type UpdateMachineProvider struct {
	SshKeyPublic *string `json:"sshKeyPublic,omitempty"`

	Aws     *UpdateMachineProviderAws     `json:"aws,omitempty"`
	Hetzner *UpdateMachineProviderHetzner `json:"hetzner,omitempty"`
}

type UpdateMachineProviderAws struct {
	AwsAccessKeyId     *string `json:"awsAccessKeyId,omitempty"`
	AwsSecretAccessKey *string `json:"awsSecretAccessKey,omitempty"`
}

type UpdateMachineProviderHetzner struct {
	CloudToken    *string `json:"cloudToken,omitempty"`
	RobotUsername *string `json:"robotUsername,omitempty"`
	RobotPassword *string `json:"robotPassword,omitempty"`
}

func MachineProviderFromDB(v dmodel.MachineProvider) *MachineProvider {
	return &MachineProvider{
		ID:            v.ID,
		Workspace:     v.WorkspaceID,
		CreatedAt:     v.CreatedAt,
		Status:        v.ReconcileStatus.ReconcileStatus.V,
		StatusDetails: v.ReconcileStatus.ReconcileStatusDetails.V,

		Type: v.Type,
		Name: v.Name,
	}
}

func MachineProviderAwsFromDB(v dmodel.MachineProviderAws) *MachineProviderAws {
	return &MachineProviderAws{
		Region:          v.Region.V,
		VpcID:           v.VpcID,
		VpcName:         v.Status.VpcName,
		VpcCidr:         v.Status.VpcCidr,
		SecurityGroupID: v.Status.SecurityGroupID,
	}
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

func MachineProviderAwsSubnetFromDB(v dmodel.MachineProviderAwsSubnet) *MachineProviderAwsSubnet {
	return &MachineProviderAwsSubnet{
		MachineProvider:  v.MachineProviderID.V,
		SubnetID:         v.SubnetID.V,
		SubnetName:       v.SubnetName,
		AvailabilityZone: v.AvailabilityZone.V,
		Cidr:             v.Cidr.V,
	}
}
