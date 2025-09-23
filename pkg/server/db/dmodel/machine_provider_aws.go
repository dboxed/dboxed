package dmodel

import (
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type MachineProviderAws struct {
	ID querier2.NullForJoin[int64] `db:"id"`

	Region             querier2.NullForJoin[string] `db:"region"`
	AwsAccessKeyID     *string                      `db:"aws_access_key_id"`
	AwsSecretAccessKey *string                      `db:"aws_secret_access_key"`
	VpcID              *string                      `db:"vpc_id"`

	Status *MachineProviderAwsStatus `join:"true"`
}

type MachineProviderAwsStatus struct {
	ID querier2.NullForJoin[int64] `db:"id"`

	VpcName         *string `db:"vpc_name"`
	VpcCidr         *string `db:"vpc_cidr"`
	SecurityGroupID *string `db:"security_group_id"`

	Subnets []MachineProviderAwsSubnet
}

type MachineProviderAwsSubnet struct {
	MachineProviderID querier2.NullForJoin[int64] `db:"machine_provider_id"`

	SubnetID         querier2.NullForJoin[string] `db:"subnet_id"`
	SubnetName       *string                      `db:"subnet_name"`
	AvailabilityZone querier2.NullForJoin[string] `db:"availability_zone"`
	Cidr             querier2.NullForJoin[string] `db:"cidr"`
}

func (v *MachineProviderAws) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *MachineProviderAwsStatus) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *MachineProviderAwsSubnet) CreateOrUpdate(q *querier2.Querier) error {
	return querier2.CreateOrUpdate(q, v, "machine_provider_id, subnet_id")
}

func DeleteMachineProviderAwsSubnet(q *querier2.Querier, machineProviderId int64, subnetId string) error {
	return querier2.DeleteOneByFields[MachineProviderAwsSubnet](q, map[string]any{
		"machine_provider_id": machineProviderId,
		"subnet_id":           subnetId,
	})
}

func GetMachineProviderSubnet(q *querier2.Querier, machineProviderId int64, subnetId string) (*MachineProviderAwsSubnet, error) {
	return querier2.GetOne[MachineProviderAwsSubnet](q, map[string]any{
		"machine_provider_id": machineProviderId,
		"subnet_id":           subnetId,
	})
}

func (v *MachineProviderAws) UpdateAccessKeys(q *querier2.Querier, awsAccessKeyID *string, awsSecretAccessKey *string) error {
	v.AwsAccessKeyID = awsAccessKeyID
	v.AwsSecretAccessKey = awsSecretAccessKey
	return querier2.UpdateOneFromStruct(q, v,
		"aws_access_key_id",
		"aws_secret_access_key",
	)
}

func (v *MachineProviderAwsStatus) UpdateVpcInfo(q *querier2.Querier, vpcName *string, vpcCidr *string) error {
	v.VpcName = vpcName
	v.VpcCidr = vpcCidr
	return querier2.UpdateOneFromStruct(q, v,
		"vpc_name",
		"vpc_cidr",
	)
}

func (v *MachineProviderAwsStatus) UpdateSecurityGroupID(q *querier2.Querier, securityGroupId *string) error {
	v.SecurityGroupID = securityGroupId
	return querier2.UpdateOneFromStruct(q, v,
		"security_group_id",
	)
}
