package dmodel

import (
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type MachineProviderHetzner struct {
	ID querier2.NullForJoin[int64] `db:"id"`

	HcloudToken        querier2.NullForJoin[string] `db:"hcloud_token"`
	RobotUser          *string                      `db:"robot_user"`
	RobotPassword      *string                      `db:"robot_password"`
	HetznerNetworkName querier2.NullForJoin[string] `db:"hetzner_network_name"`

	Status *MachineProviderHetznerStatus `join:"true"`
}

type MachineProviderHetznerStatus struct {
	ID querier2.NullForJoin[int64] `db:"id"`

	HetznerNetworkID   *int64  `db:"hetzner_network_id"`
	HetznerNetworkZone *string `db:"hetzner_network_zone"`
	HetznerNetworkCidr *string `db:"hetzner_network_cidr"`
	CloudSubnetCidr    *string `db:"cloud_subnet_cidr"`
	RobotSubnetCidr    *string `db:"robot_subnet_cidr"`
	RobotVswitchID     *int64  `db:"robot_vswitch_id"`
}

func (v *MachineProviderHetzner) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *MachineProviderHetznerStatus) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *MachineProviderHetzner) UpdateHCloudToken(q *querier2.Querier, token string) error {
	v.HcloudToken = querier2.N(token)
	return querier2.UpdateOneFromStruct(q, v, "hcloud_token")
}

func (v *MachineProviderHetzner) UpdateRobotCredentials(q *querier2.Querier, username *string, password *string) error {
	v.RobotUser = username
	v.RobotPassword = password
	return querier2.UpdateOneFromStruct(q, v,
		"robot_user",
		"robot_password",
	)
}

func (v *MachineProviderHetznerStatus) UpdateStatus(q *querier2.Querier) error {
	return querier2.UpdateOneFromStruct(q, v,
		"hetzner_network_id",
		"hetzner_network_zone",
		"hetzner_network_cidr",
		"cloud_subnet_cidr",
		"robot_subnet_cidr",
		"robot_vswitch_id",
	)
}
