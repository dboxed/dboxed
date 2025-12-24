package dmodel

import (
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type DboxedSpecMapping struct {
	OwnedByWorkspace

	SpecId     string `db:"spec_id"`
	ObjectType string `db:"object_type"`
	ObjectId   string `db:"object_id"`
	ObjectName string `db:"object_name"`

	RecreateKey  string `db:"recreate_key"`
	SpecFragment string `db:"spec_fragment"`
}

func (v *DboxedSpecMapping) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func GetDboxedSpecMappingById(q *querier2.Querier, workspaceId *string, id string) (*DboxedSpecMapping, error) {
	return querier2.GetOne[DboxedSpecMapping](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           id,
	})
}

func ListDboxedSpecMappingForSpec(q *querier2.Querier, workspaceId string, specId string) ([]DboxedSpecMapping, error) {
	return querier2.GetMany[DboxedSpecMapping](q, map[string]any{
		"workspace_id": workspaceId,
		"spec_id":      specId,
	}, nil)
}

func (v *DboxedSpecMapping) UpdateSpecFragment(q *querier2.Querier, specFragment string) error {
	v.SpecFragment = specFragment
	return querier2.UpdateOneFromStruct(q, v, "spec_fragment")
}
