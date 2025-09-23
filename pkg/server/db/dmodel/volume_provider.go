package dmodel

import (
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type VolumeProvider struct {
	OwnedByWorkspace
	ReconcileStatus

	Type string `db:"type"`
	Name string `db:"name"`

	Dboxed *VolumeProviderDboxed `join:"true"`
}

func (v *VolumeProvider) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func GetVolumeProviderById(q *querier2.Querier, workspaceId *int64, id int64, skipDeleted bool) (*VolumeProvider, error) {
	v, err := querier2.GetOne[VolumeProvider](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           id,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
	if err != nil {
		return nil, err
	}
	return v, nil
}

func ListVolumeProviders(q *querier2.Querier, workspaceId int64, skipDeleted bool) ([]VolumeProvider, error) {
	l, err := querier2.GetMany[VolumeProvider](q, map[string]any{
		"workspace_id": workspaceId,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
	if err != nil {
		return nil, err
	}
	return l, nil
}
