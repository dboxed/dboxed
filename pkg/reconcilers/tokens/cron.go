package tokens

import (
	"context"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/util"
)

type CronJob struct {
}

func NewCronJob() *CronJob {
	return &CronJob{}
}

func (r *CronJob) Run(ctx context.Context) error {
	log := slog.With("job", "tokens")

	for {
		err := r.runOnce(ctx, log)
		if err != nil {
			log.ErrorContext(ctx, "error in runOnce", "error", err)
		}

		if !util.SleepWithContext(ctx, time.Second*60) {
			return ctx.Err()
		}
	}
}

func (r *CronJob) runOnce(ctx context.Context, log *slog.Logger) error {
	q := querier.GetQuerier(ctx)

	expiredTokens, err := dmodel.ListExpiredInternalTokens(q, time.Now())
	if err != nil {
		return err
	}

	for _, t := range expiredTokens {
		log.InfoContext(ctx, "deleting expired token", "workspaceId", t.WorkspaceID, "tokenId", t.ID, "tokenName", t.Name)
		err = querier.DeleteOneById[dmodel.Token](q, t.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
