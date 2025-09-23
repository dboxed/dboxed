package dmodel

import (
	"fmt"
	"strings"

	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type Workspace struct {
	ID int64 `db:"id" omitCreate:"true"`
	SoftDeleteFields
	Times

	ReconcileStatus

	Name     string `db:"name"`
	Nkey     string `db:"nkey"`
	NkeySeed string `db:"nkey_seed"`

	Access []WorkspaceAccess
}

type WorkspaceAccess struct {
	WorkspaceId int64  `db:"workspace_id"`
	UserId      string `db:"user_id"`

	User User `join:"true" join_left_field:"user_id"`
}

func (v *Workspace) SetId(id int64) {
	v.ID = id
}

func (v Workspace) GetId() int64 {
	return v.ID
}

func (v *Workspace) Create(q *querier2.Querier) error {
	err := querier2.Create(q, v)
	if err != nil {
		return err
	}

	for i, wa := range v.Access {
		wa.WorkspaceId = v.ID
		err = wa.Create(q)
		if err != nil {
			return err
		}
		v.Access[i] = wa
	}
	return nil
}

func (v *WorkspaceAccess) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func GetWorkspaceAccessesById(q *querier2.Querier, id int64) ([]WorkspaceAccess, error) {
	l, err := querier2.GetMany[WorkspaceAccess](q, map[string]any{
		"workspace_id": id,
	})
	if err != nil {
		return nil, err
	}
	return l, nil
}

func postprocessWorkspace(q *querier2.Querier, w *Workspace) (*Workspace, error) {
	was, err := GetWorkspaceAccessesById(q, w.ID)
	if err != nil {
		return nil, err
	}

	w.Access = was
	return w, nil
}

func ListWorkspaces(q *querier2.Querier, userId *string, skipDeleted bool) ([]Workspace, error) {
	wasWhere, wasWhereArgs, err := querier2.BuildWhere[WorkspaceAccess](map[string]any{
		"user_id": querier2.OmitIfNull(userId),
	})
	if err != nil {
		return nil, err
	}

	was, err := querier2.GetManyWhere[WorkspaceAccess](q, wasWhere, wasWhereArgs)
	if err != nil {
		return nil, err
	}
	wasMap := map[int64][]WorkspaceAccess{}
	for _, wa := range was {
		wasMap[wa.WorkspaceId] = append(wasMap[wa.WorkspaceId], wa)
	}

	if wasWhere != "" {
		wasWhere = "where " + wasWhere
	}
	var whereClauses []string
	whereClauses = append(whereClauses, fmt.Sprintf("id in (select workspace_id from workspace_access %s)", wasWhere))
	if skipDeleted {
		whereClauses = append(whereClauses, "deleted_at is null")
	}
	where := strings.Join(whereClauses, " and ")
	l, err := querier2.GetManyWhere[Workspace](q, where, wasWhereArgs)
	if err != nil {
		return nil, err
	}

	for i, x := range l {
		l[i].Access = wasMap[x.ID]
	}
	return l, nil
}

func GetWorkspaceById(q *querier2.Querier, id int64, skipDeleted bool) (*Workspace, error) {
	w, err := querier2.GetOne[Workspace](q, map[string]any{
		"id":         id,
		"deleted_at": querier2.ExcludeNonNull(skipDeleted),
	})
	if err != nil {
		return nil, err
	}
	return postprocessWorkspace(q, w)
}

func GetWorkspaceByNkey(q *querier2.Querier, nkey string) (*Workspace, error) {
	w, err := querier2.GetOne[Workspace](q, map[string]any{
		"nkey": nkey,
	})
	if err != nil {
		return nil, err
	}
	return postprocessWorkspace(q, w)
}
