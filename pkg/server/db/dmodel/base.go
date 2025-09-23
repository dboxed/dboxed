package dmodel

import (
	"fmt"
	"time"

	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type HasReconcileStatus interface {
	querier2.HasId
	IsSoftDelete
	SetReconcileStatus(status string, statusDetails string)
	GetReconcileStatus() (string, string)
}

type Times struct {
	CreatedAt time.Time `db:"created_at" omitCreate:"true"`
}

type OwnedByWorkspace struct {
	ID int64 `db:"id" omitCreate:"true"`
	SoftDeleteFields
	Times

	WorkspaceID int64 `db:"workspace_id"`
	Workspace   *Workspace
}

type ReconcileStatus struct {
	ReconcileStatus        string `db:"reconcile_status" omitCreate:"true"`
	ReconcileStatusDetails string `db:"reconcile_status_details" omitCreate:"true"`
}

func (v *OwnedByWorkspace) SetId(id int64) {
	v.ID = id
}

func (v OwnedByWorkspace) GetId() int64 {
	return v.ID
}

func (v *ReconcileStatus) SetReconcileStatus(status string, statusDetails string) {
	v.ReconcileStatus = status
	v.ReconcileStatusDetails = statusDetails
}

func (v *ReconcileStatus) GetReconcileStatus() (string, string) {
	return v.ReconcileStatus, v.ReconcileStatusDetails
}

func UpdateReconcileStatus[T HasReconcileStatus](q *querier2.Querier, v T) error {
	return querier2.UpdateOneFromStruct(q, &v, "reconcile_status", "reconcile_status_details")
}

func GetAllIds[T querier2.HasId](q *querier2.Querier) ([]int64, error) {
	var ret []int64
	err := q.SelectNamed(&ret, fmt.Sprintf("select id from %s", querier2.GetTableName[T]()), nil)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
