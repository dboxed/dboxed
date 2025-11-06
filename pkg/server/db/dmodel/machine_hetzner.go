package dmodel

import (
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type MachineHetzner struct {
	ID querier2.NullForJoin[string] `db:"id"`
	ReconcileStatus

	ServerType     querier2.NullForJoin[string] `db:"server_type"`
	ServerLocation querier2.NullForJoin[string] `db:"server_location"`

	Status *MachineHetznerStatus `join:"true"`
}

func (v *MachineHetzner) GetId() string {
	return v.ID.V
}

type MachineHetznerStatus struct {
	ID querier2.NullForJoin[string] `db:"id"`

	ServerID *int64 `db:"server_id"`
}

func (v *MachineHetzner) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *MachineHetznerStatus) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *MachineHetznerStatus) UpdateServerID(q *querier2.Querier, serverId *int64) error {
	v.ServerID = serverId
	return querier2.UpdateOneFromStruct(q, v, "server_id")
}
