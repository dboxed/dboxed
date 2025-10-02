package dbutils

import (
	"context"

	"github.com/dboxed/dboxed/pkg/server/db/querier"
)

func RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	db := querier.GetDB(ctx)

	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	doRollback := true
	defer func() {
		if doRollback {
			_ = tx.Rollback()
		}
	}()
	ctx = context.WithValue(ctx, "tx", tx)

	err = fn(ctx)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	doRollback = false
	return nil
}
