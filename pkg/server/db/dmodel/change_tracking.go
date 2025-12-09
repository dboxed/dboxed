package dmodel

import (
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

func BumpChangeSeq[T HasReconcileStatus](q *querier2.Querier, v T) error {
	return BumpChangeSeqForId[T](q, v.GetId())
}

func BumpChangeSeqForId[T HasReconcileStatus](q *querier2.Querier, id string) error {
	return querier2.UpdateOne[T](q, "id = :id", map[string]any{
		"id": id,
	}, map[string]any{
		"change_seq": querier2.RawSql("nextval('change_tracking_seq')"),
	})
}

func GetMaxChangeSeq[T HasReconcileStatus](q *querier2.Querier) (int64, error) {
	var maxSeq int64
	err := q.GetNamed(&maxSeq, "select coalesce(max(change_seq), -1) from "+querier2.GetTableName[T](), nil)
	if err != nil {
		return 0, err
	}
	return maxSeq, nil
}

type IdAndChangeSeq struct {
	Id        string `db:"id"`
	ChangeSeq int64  `db:"change_seq"`
}

func FindChanges[T HasReconcileStatus](q *querier2.Querier, lastSeq int64) ([]IdAndChangeSeq, error) {
	var ret []IdAndChangeSeq
	err := q.SelectNamed(&ret, "select id, change_seq from "+querier2.GetTableName[T]()+" where change_seq > :last_seq", map[string]any{
		"last_seq": lastSeq,
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}
