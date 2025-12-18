package base

import (
	"context"

	"github.com/dboxed/dboxed/pkg/server/db/querier"
)

func Transaction(ctx context.Context, fn func(ctx context.Context) ReconcileResult) ReconcileResult {
	var result ReconcileResult
	err := querier.Transaction(ctx, func(ctx context.Context) (bool, error) {
		result = fn(ctx)
		if result.Error != nil {
			return false, result.Error
		}
		return true, nil
	})
	if result.Error == nil && err != nil {
		return InternalError(err)
	}
	return result
}
