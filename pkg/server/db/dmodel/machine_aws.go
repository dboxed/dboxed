package dmodel

import (
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type MachineAws struct {
	ID querier2.NullForJoin[int64] `db:"id"`

	InstanceType   querier2.NullForJoin[string] `db:"instance_type"`
	SubnetID       querier2.NullForJoin[string] `db:"subnet_id"`
	RootVolumeSize querier2.NullForJoin[int64]  `db:"root_volume_size"`

	Status *MachineAwsStatus `join:"true"`
}

type MachineAwsStatus struct {
	ID querier2.NullForJoin[int64] `db:"id"`

	InstanceID *string `db:"instance_id"`
}

func (v *MachineAws) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *MachineAwsStatus) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *MachineAwsStatus) UpdateInstanceID(q *querier2.Querier, instanceId *string) error {
	v.InstanceID = instanceId
	return querier2.UpdateOneFromStruct(q, v, "instance_id")
}
