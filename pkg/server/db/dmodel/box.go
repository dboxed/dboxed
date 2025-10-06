package dmodel

import (
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type Box struct {
	OwnedByWorkspace
	ReconcileStatus

	Uuid string `db:"uuid"`
	Name string `db:"name"`

	NetworkID   *int64  `db:"network_id"`
	NetworkType *string `db:"network_type"`

	DboxedVersion string `db:"dboxed_version"`
	BoxSpec       []byte `db:"box_spec"`

	MachineID *int64 `db:"machine_id"`

	Netbird *BoxNetbird `join:"true"`
}

type BoxNetbird struct {
	ID querier2.NullForJoin[int64] `db:"id"`

	SetupKey   *string `db:"setup_key"`
	SetupKeyID *string `db:"setup_key_id"`
}

func (v *Box) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *BoxNetbird) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func GetBoxById(q *querier2.Querier, workspaceId *int64, id int64, skipDeleted bool) (*Box, error) {
	return querier2.GetOne[Box](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           id,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func GetBoxByName(q *querier2.Querier, workspaceId int64, name string, skipDeleted bool) (*Box, error) {
	return querier2.GetOne[Box](q, map[string]any{
		"workspace_id": workspaceId,
		"name":         name,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func GetBoxByUuid(q *querier2.Querier, workspaceId int64, uuid string, skipDeleted bool) (*Box, error) {
	return querier2.GetOne[Box](q, map[string]any{
		"workspace_id": workspaceId,
		"uuid":         uuid,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func ListBoxesForWorkspace(q *querier2.Querier, workspaceId int64, skipDeleted bool) ([]Box, error) {
	return querier2.GetMany[Box](q, map[string]any{
		"workspace_id": workspaceId,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func ListBoxesForNetwork(q *querier2.Querier, networkId int64, skipDeleted bool) ([]Box, error) {
	return querier2.GetMany[Box](q, map[string]any{
		"network_id": networkId,
		"deleted_at": querier2.ExcludeNonNull(skipDeleted),
	})
}

func (v *Box) UpdateBoxSpec(q *querier2.Querier, b []byte) error {
	v.BoxSpec = b
	return querier2.UpdateOneFromStruct(q, v, "box_spec")
}

func (v *BoxNetbird) UpdateSetupKey(q *querier2.Querier, setupKey *string, setupKeyId *string) error {
	v.SetupKey = setupKey
	v.SetupKeyID = setupKeyId
	return querier2.UpdateOneFromStruct(q, v,
		"setup_key",
		"setup_key_id",
	)
}
