package dmodel

import (
	"time"

	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type ChangeTracking struct {
	ID int64 `db:"id" omitCreate:"true"`

	TableName string    `db:"table_name"`
	EntityID  int64     `db:"entity_id"`
	Time      time.Time `db:"time" omitCreate:"true"`
}

func AddChangeTracking[T querier2.HasId](q *querier2.Querier, v T) error {
	return AddChangeTrackingForId[T](q, v.GetId())
}

func AddChangeTrackingForId[T querier2.HasId](q *querier2.Querier, id int64) error {
	return querier2.Create(q, &ChangeTracking{
		TableName: querier2.GetTableName[T](),
		EntityID:  id,
	})
}

const queryGetMaxChangeTrackingId = `select coalesce(max(id), -1) from change_tracking where table_name = :table_name`

func GetMaxChangeTrackingId[T querier2.HasId](q *querier2.Querier) (int64, error) {
	var maxId int64
	err := q.GetNamed(&maxId, queryGetMaxChangeTrackingId, ChangeTracking{
		TableName: querier2.GetTableName[T](),
	})
	if err != nil {
		return 0, err
	}
	return maxId, nil
}

func FindChanges[T querier2.HasId](q *querier2.Querier, lastId int64) ([]ChangeTracking, error) {
	return querier2.GetManyWhere[ChangeTracking](q, "table_name = :table_name and id > :last_id", map[string]any{
		"table_name": querier2.GetTableName[T](),
		"last_id":    lastId,
	})
}
