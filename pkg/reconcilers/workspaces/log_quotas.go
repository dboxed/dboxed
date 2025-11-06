package workspaces

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/jmoiron/sqlx"
)

func (r *reconciler) reconcileLogQuotas(ctx context.Context, w *dmodel.Workspace, log *slog.Logger) base.ReconcileResult {
	q := querier.GetQuerier(ctx)
	db := querier.GetDB(ctx)

	wq, err := dmodel.GetWorkspaceQuotaById(q, w.ID)
	if err != nil {
		return base.InternalError(err)
	}

	wu, err := dmodel.QueryWorkspaceLogBytesUsage(q, w.ID)
	if err != nil {
		if querier.IsSqlNotFoundError(err) {
			return base.ReconcileResult{}
		}
		return base.InternalError(err)
	}

	querySize := int64(100)
	maxQuerySize := int64(10000)

	logMetadataSubs := map[string]int64{}
	for wu.SumLineBytes > wq.MaxLogBytes {
		clear(logMetadataSubs)
		deleteUntilId := int64(-1)
		deleteLineCount := 0
		deleteLineBytes := int64(0)
		llbs, err := dmodel.QueryLogLineBytes(q, w.ID, querySize)
		if err != nil {
			return base.InternalError(err)
		}
		querySize = min(querySize*2, maxQuerySize)
		for _, llb := range llbs {
			wu.SumLineBytes -= llb.LineBytes
			deleteLineBytes += llb.LineBytes
			logMetadataSubs[llb.LogID] += llb.LineBytes
			deleteUntilId = llb.ID
			deleteLineCount++
			if wu.SumLineBytes <= wq.MaxLogBytes {
				break
			}
		}
		if deleteUntilId != -1 {
			var deletedLines int64
			err := db.Transaction(ctx, func(tx *sqlx.Tx) (bool, error) {
				var err error
				q := querier.NewQuerier(ctx, db, tx)
				deletedLines, err = dmodel.DeleteLogLinesUntilId(q, w.ID, deleteUntilId)
				if err != nil {
					return false, err
				}
				for logId, sub := range logMetadataSubs {
					err = dmodel.AddLogMetadataTotalBytes(q, logId, -sub)
					if err != nil {
						return false, err
					}
				}

				return true, nil
			})
			if err != nil {
				return base.InternalError(err)
			}

			log.InfoContext(ctx, "deleted log lines",
				slog.Any("deleteUntilId", deleteUntilId),
				slog.Any("deletedLines", deletedLines),
				slog.Any("sumLineBytes", wu.SumLineBytes),
				slog.Any("deletedLineBytes", deleteLineBytes),
				slog.Any("deleteLineCount", deleteLineCount),
			)
		}
	}

	return base.ReconcileResult{}
}
