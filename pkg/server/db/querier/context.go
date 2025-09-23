package querier

import (
	"context"

	"github.com/jmoiron/sqlx"
)

func GetDB(c context.Context) *sqlx.DB {
	i := c.Value("db")
	if i == nil {
		panic("context has no db")
	}
	db, ok := i.(*sqlx.DB)
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

func GetTX(c context.Context) *sqlx.Tx {
	return getTX(c, true)
}

func GetQuerier(c context.Context) *Querier {
	tx := getTX(c, false)
	var tx2 *sqlx.Tx
	if tx != nil {
		tx2 = tx
	}
	return NewQuerier(c, GetDB(c), tx2)
}
