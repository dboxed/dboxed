package querier

import (
	"context"

	"github.com/jmoiron/sqlx"
)

func GetDB(c context.Context) *ReadWriteDB {
	i := c.Value("db")
	if i == nil {
		panic("context has no db")
	}
	db, ok := i.(*ReadWriteDB)
	if !ok {
		panic("db in context has wrong type")
	}
	return db
}

func getTX(c context.Context, doPanic bool) *sqlx.Tx {
	i := c.Value("tx")
	if i == nil {
		if !doPanic {
			return nil
		}
		panic("context has no tx")
	}
	tx, ok := i.(*sqlx.Tx)
	if !ok {
		if !doPanic {
			return nil
		}
		panic("tx in context has wrong type")
	}
	return tx
}

func GetQuerier(c context.Context) *Querier {
	tx := getTX(c, false)
	var tx2 *sqlx.Tx
	if tx != nil {
		tx2 = tx
	}
	return NewQuerier(c, GetDB(c), tx2)
}

func WithForbidAutoTx(ctx context.Context) context.Context {
	return context.WithValue(ctx, "forbid-auto-tx", true)
}

func IsForbidAutoTx(ctx context.Context) bool {
	v := ctx.Value("forbid-auto-tx")
	if v == nil {
		return false
	}
	b, ok := v.(bool)
	if !ok || !b {
		return false
	}
	return true
}

func Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return GetDB(ctx).Transaction(ctx, func(tx *sqlx.Tx) (bool, error) {
		ctx = context.WithValue(ctx, "tx", tx)
		err := fn(ctx)
		if err != nil {
			return false, err
		}
		return true, nil
	})
}
