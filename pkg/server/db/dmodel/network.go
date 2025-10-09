package dmodel

import (
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type Network struct {
	OwnedByWorkspace
	ReconcileStatus

	Type string `db:"type"`
	Name string `db:"name"`

	Netbird *NetworkNetbird `join:"true"`
}

func (v *Network) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func GetNetworkById(q *querier2.Querier, workspaceId *int64, id int64, skipDeleted bool) (*Network, error) {
	return querier2.GetOne[Network](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           id,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func ListNetworksForWorkspace(q *querier2.Querier, workspaceId int64, skipDeleted bool) ([]Network, error) {
	return querier2.GetMany[Network](q, map[string]any{
		"workspace_id": workspaceId,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	}, nil)
}
