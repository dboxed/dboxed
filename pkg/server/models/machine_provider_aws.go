package models

import "github.com/dboxed/dboxed/pkg/server/db/dmodel"

type MachineProviderAws struct {
	Region          string  `json:"region"`
	VpcID           *string `json:"vpcId"`
	VpcName         *string `json:"vpcName"`
	VpcCidr         *string `json:"vpcCidr"`
	SecurityGroupID *string `json:"securityGroupId"`

	Subnets []MachineProviderAwsSubnet `json:"subnets"`
}

type MachineProviderAwsSubnet struct {
	MachineProvider  string  `json:"machineProvider"`
	SubnetID         string  `json:"subnetId"`
	SubnetName       *string `json:"subnetName"`
	AvailabilityZone string  `json:"availabilityZone"`
	Cidr             string  `json:"cidr"`
}

type CreateMachineProviderAws struct {
	Region string `json:"region"`
	VpcId  string `json:"vpcId"`

	AwsAccessKeyId     string `json:"awsAccessKeyId"`
	AwsSecretAccessKey string `json:"awsSecretAccessKey"`
}

type UpdateMachineProviderAws struct {
	AwsAccessKeyId     *string `json:"awsAccessKeyId,omitempty"`
	AwsSecretAccessKey *string `json:"awsSecretAccessKey,omitempty"`
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

func MachineProviderAwsSubnetFromDB(v dmodel.MachineProviderAwsSubnet) *MachineProviderAwsSubnet {
	return &MachineProviderAwsSubnet{
		MachineProvider:  v.MachineProviderID.V,
		SubnetID:         v.SubnetID.V,
		SubnetName:       v.SubnetName,
		AvailabilityZone: v.AvailabilityZone.V,
		Cidr:             v.Cidr.V,
	}
}
