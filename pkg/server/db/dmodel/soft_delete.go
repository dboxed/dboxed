package dmodel

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/google/uuid"
)

type IsSoftDelete interface {
	SetId(id string)
	GetId() string
	GetDeletedAt() *time.Time
	SetDeletedAt(t *time.Time)
	GetFinalizers() []string
	SetFinalizers(finalizers []string)
	HasFinalizer(k string) bool

	setFinalizersRaw(finalizers string)
}

type SoftDeleteFields struct {
	DeletedAt  sql.NullTime `db:"deleted_at" omitCreate:"true"`
	Finalizers string       `db:"finalizers" omitCreate:"true"`
}

func (v *SoftDeleteFields) GetDeletedAt() *time.Time {
	if !v.DeletedAt.Valid {
		return nil
	}
	return &v.DeletedAt.Time
}

func (v *SoftDeleteFields) SetDeletedAt(t *time.Time) {
	if t == nil {
		v.DeletedAt = sql.NullTime{}
	} else {
		v.DeletedAt = sql.NullTime{
			Valid: true,
			Time:  *t,
		}
	}
}

func (v *SoftDeleteFields) GetFinalizers() []string {
	if v.Finalizers == "{}" || v.Finalizers == "" {
		return nil
	}
	var m map[string]any
	err := json.Unmarshal([]byte(v.Finalizers), &m)
	if err != nil {
		panic(err)
	}
	ret := make([]string, 0, len(m))
	for k := range m {
		ret = append(ret, k)
	}
	return ret
}

func (v *SoftDeleteFields) SetFinalizers(finalizers []string) {
	m := map[string]bool{}
	for _, x := range finalizers {
		m[x] = true
	}
	v.Finalizers = util.MustJson(m)
}

func (v *SoftDeleteFields) setFinalizersRaw(finalizers string) {
	v.Finalizers = finalizers
}

func (v *SoftDeleteFields) HasFinalizer(k string) bool {
	return slices.Contains(v.GetFinalizers(), k)
}

func SoftDelete[T IsSoftDelete](q *querier2.Querier, byFields map[string]any) error {
	return querier2.UpdateOneByFields[T](q, byFields, map[string]any{
		"deleted_at": querier2.RawSql("current_timestamp"),
	})
}

func SoftDeleteByIds[T IsSoftDelete](q *querier2.Querier, workspaceId *string, id string) error {
	err := SoftDelete[T](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           id,
	})
	if err != nil {
		return err
	}
	return nil
}

func SoftDeleteByStruct[T IsSoftDelete](q *querier2.Querier, v T) error {
	err := SoftDeleteByIds[T](q, nil, v.GetId())
	if err != nil {
		return err
	}
	l, err := querier2.GetMany[T](q, map[string]any{
		"id": v.GetId(),
	}, nil)
	if err != nil {
		return err
	}
	v.SetDeletedAt(l[0].GetDeletedAt())
	return nil
}

type SoftDeleteWithConstraintsExtra func(q *querier2.Querier) error

func SoftDeleteWithConstraints[T IsSoftDelete](q *querier2.Querier, byFields map[string]any, extra SoftDeleteWithConstraintsExtra) error {
	savepoint := "s_" + strings.ReplaceAll(uuid.NewString(), "-", "_")
	_, err := q.ExecNamed(fmt.Sprintf("savepoint %s", savepoint), nil)
	if err != nil {
		return err
	}

	if extra != nil {
		err = extra(q)
		if err != nil {
			return err
		}
	}

	err = querier2.DeleteOneByFields[T](q, byFields)
	if err != nil {
		return err
	}
	_, err = q.ExecNamed(fmt.Sprintf("rollback to savepoint %s", savepoint), nil)
	if err != nil {
		return err
	}

	err = SoftDelete[T](q, byFields)
	if err != nil {
		return err
	}

	return nil
}

func SoftDeleteWithConstraintsByIdsExtra[T IsSoftDelete](q *querier2.Querier, workspaceId *string, id string, extra SoftDeleteWithConstraintsExtra) error {
	err := SoftDeleteWithConstraints[T](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           id,
	}, extra)
	if err != nil {
		return err
	}
	err = AddChangeTrackingForId[T](q, id)
	if err != nil {
		return err
	}
	return nil
}

func SoftDeleteWithConstraintsByIds[T IsSoftDelete](q *querier2.Querier, workspaceId *string, id string) error {
	return SoftDeleteWithConstraintsByIdsExtra[T](q, workspaceId, id, nil)
}

var querySetDBFinalizers = map[string]string{
	"pgx": `update @@table_name
set    finalizers = jsonb_strip_nulls(jsonb_set(to_jsonb(finalizers::::json), '{@@k}', '@@nullOrTrue'))
where id = :id
returning finalizers`,
	"sqlite3": `
update @@table_name
set    finalizers = json_patch(finalizers, '{"@@k":: @@nullOrTrue}')
where id = :id
returning finalizers`,
}

func setDBFinalizers[T any](q *querier2.Querier, id string, k string, v bool) (string, error) {
	nullOrTrue := "null"
	if v {
		nullOrTrue = "true"
	}

	var newFinalizers string
	err := q.GetNamed(&newFinalizers, querySetDBFinalizers, map[string]any{
		"id":           id,
		"@@table_name": querier2.GetTableName[T](),
		"@@k":          k,
		"@@nullOrTrue": nullOrTrue,
	})
	if err != nil {
		return "", err
	}

	return newFinalizers, nil
}

func AddFinalizer[T IsSoftDelete](q *querier2.Querier, v T, finalizer string) error {
	if v.HasFinalizer(finalizer) {
		return nil
	}

	newFinalizers, err := setDBFinalizers[T](q, v.GetId(), finalizer, true)
	if err != nil {
		return err
	}

	v.setFinalizersRaw(newFinalizers)

	return nil
}

func RemoveFinalizer[T IsSoftDelete](q *querier2.Querier, v T, finalizer string) error {
	if !v.HasFinalizer(finalizer) {
		return nil
	}

	newFinalizers, err := setDBFinalizers[T](q, v.GetId(), finalizer, false)
	if err != nil {
		return err
	}

	v.setFinalizersRaw(newFinalizers)

	return nil
}
