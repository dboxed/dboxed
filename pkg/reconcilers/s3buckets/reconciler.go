package s3buckets

import (
	"context"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
)

type reconciler struct {
}

func NewS3BucketsReconciler(config config.Config) *base.Reconciler[*dmodel.S3Bucket] {
	return base.NewReconciler(base.Config[*dmodel.S3Bucket]{
		ServerConfig:          config,
		ReconcilerName:        "s3buckets",
		Impl:                  &reconciler{},
		FullReconcileInterval: 10 * time.Second,
	})
}

func (r *reconciler) GetItem(ctx context.Context, id int64) (*dmodel.S3Bucket, error) {
	return dmodel.GetS3BucketById(querier.GetQuerier(ctx), nil, id, false)
}

func (r *reconciler) Reconcile(ctx context.Context, s *dmodel.S3Bucket, log *slog.Logger) error {
	log = log.With(
		slog.Any("bucket", s.Bucket),
		slog.Any("endpoint", s.Endpoint),
	)

	// Nothing to reconcile for now - S3 buckets don't need active management
	// Future improvements could include:
	// - Validating S3 credentials
	// - Checking bucket accessibility
	// - Monitoring bucket usage

	return nil
}

func (r *reconciler) ReconcileDelete(ctx context.Context, s *dmodel.S3Bucket, log *slog.Logger) error {
	log = log.With(
		slog.Any("bucket", s.Bucket),
		slog.Any("endpoint", s.Endpoint),
	)

	// For now, we only handle final deletion
	// We don't delete the actual S3 bucket or its contents
	// The bucket configuration is simply removed from the database
	// after soft-delete (when finalizers are cleared)

	log.InfoContext(ctx, "s3 bucket configuration ready for final deletion")

	return nil
}
