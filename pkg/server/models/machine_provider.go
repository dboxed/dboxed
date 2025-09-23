package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/global"
)

type MachineProvider struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Workspace int64     `json:"workspace"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`

	SshKeyFingerprint *string `json:"ssh_key_fingerprint"`

	Aws     *MachineProviderAws     `json:"aws,omitempty"`
	Hetzner *MachineProviderHetzner `json:"hetzner,omitempty"`
}

type MachineProviderAws struct {
	Region          string  `json:"region"`
	VpcID           *string `json:"vpc_id"`
	VpcName         *string `json:"vpc_name"`
	VpcCidr         *string `json:"vpc_cidr"`
	SecurityGroupID *string `json:"security_group_id"`

	Subnets []MachineProviderAwsSubnet `json:"subnets"`
}

type MachineProviderAwsSubnet struct {
	MachineProvider  int64   `json:"machine_provider"`
	SubnetID         string  `json:"subnet_id"`
	SubnetName       *string `json:"subnet_name"`
	AvailabilityZone string  `json:"availability_zone"`
	Cidr             string  `json:"cidr"`
}

type MachineProviderHetzner struct {
	HetznerNetworkName string  `json:"hetzner_network_name"`
	HetznerNetworkID   *int64  `json:"hetzner_network_id"`
	HetznerNetworkZone *string `json:"hetzner_network_zone"`
	HetznerNetworkCidr *string `json:"hetzner_network_cidr"`
	CloudSubnetCidr    *string `json:"cloud_subnet_cidr"`
	RobotSubnetCidr    *string `json:"robot_subnet_cidr"`
	RobotVswitchID     *int64  `json:"robot_vswitch_id"`
}

type CreateMachineProvider struct {
	Type global.MachineProviderType `json:"type"`
	Name string                     `json:"name"`

	SshKeyPublic *string `json:"ssh_key_public,omitempty"`

	Aws     *CreateMachineProviderAws     `json:"aws,omitempty"`
	Hetzner *CreateMachineProviderHetzner `json:"hetzner,omitempty"`
}

type CreateMachineProviderAws struct {
	Region string `json:"region"`
	VpcId  string `json:"vpc_id"`

	AwsAccessKeyId     string `json:"aws_access_key_id"`
	AwsSecretAccessKey string `json:"aws_secret_access_key"`
}

type CreateMachineProviderHetzner struct {
	CloudToken    string  `json:"cloud_token"`
	RobotUsername *string `json:"robot_username,omitempty"`
	RobotPassword *string `json:"robot_password,omitempty"`

	HetznerNetworkName string `json:"hetzner_network_name"`
}

type UpdateMachineProvider struct {
	SshKeyPublic *string `json:"ssh_key_public,omitempty"`

	Aws     *UpdateMachineProviderAws     `json:"aws,omitempty"`
	Hetzner *UpdateMachineProviderHetzner `json:"hetzner,omitempty"`
}

type UpdateMachineProviderAws struct {
	AwsAccessKeyId     *string `json:"aws_access_key_id,omitempty"`
	AwsSecretAccessKey *string `json:"aws_secret_access_key,omitempty"`
}

type UpdateMachineProviderHetzner struct {
	CloudToken    *string `json:"cloud_token,omitempty"`
	RobotUsername *string `json:"robot_username,omitempty"`
	RobotPassword *string `json:"robot_password,omitempty"`
}

func MachineProviderFromDB(v dmodel.MachineProvider) *MachineProvider {
	return &MachineProvider{
		ID:        v.ID,
		Workspace: v.WorkspaceID,
		CreatedAt: v.CreatedAt,
		Type:      v.Type,
		Name:      v.Name,
		Status:    v.ReconcileStatus.ReconcileStatus,
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
