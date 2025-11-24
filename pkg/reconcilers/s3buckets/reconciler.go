package s3buckets

import (
	"context"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
)

type reconciler struct {
}

func NewS3BucketsReconciler() *base.Reconciler[*dmodel.S3Bucket] {
	return base.NewReconciler(base.Config[*dmodel.S3Bucket]{
		ReconcilerName:        "s3buckets",
		Impl:                  &reconciler{},
		FullReconcileInterval: 10 * time.Second,
	})
}

func (r *reconciler) GetItem(ctx context.Context, id string) (*dmodel.S3Bucket, error) {
	return dmodel.GetS3BucketById(querier.GetQuerier(ctx), nil, id, false)
}

func (r *reconciler) Reconcile(ctx context.Context, s *dmodel.S3Bucket, log *slog.Logger) base.ReconcileResult {
	log = log.With(
		slog.Any("bucket", s.Bucket),
		slog.Any("endpoint", s.Endpoint),
	)

	// Nothing to reconcile for now - S3 buckets don't need active management
	// Future improvements could include:
	// - Validating S3 credentials
	// - Checking bucket accessibility
	// - Monitoring bucket usage

	return base.ReconcileResult{}
}
