package s3buckets

import (
	"context"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/s3utils"
)

type reconciler struct {
}

func NewS3BucketsReconciler() *base.Reconciler[*dmodel.S3Bucket] {
	return base.NewReconciler(base.Config[*dmodel.S3Bucket]{
		ReconcilerName:        "s3buckets",
		Reconciler:            &reconciler{},
		FullReconcileInterval: 60 * time.Second,
	})
}

func (r *reconciler) GetItem(ctx context.Context, id string) (*dmodel.S3Bucket, error) {
	return dmodel.GetS3BucketById(querier.GetQuerier(ctx), nil, id, false)
}

func (r *reconciler) Reconcile(ctx context.Context, s *dmodel.S3Bucket, log *slog.Logger) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	log = log.With(
		slog.Any("bucket", s.Bucket),
		slog.Any("endpoint", s.Endpoint),
	)

	// TODO remove this code after some time. Locations are pre-determined on S3 bucket creation and
	// this here is only for compatibility
	if s.DeterminedRegion == nil {
		c, err := s3utils.BuildS3Client(s)
		if err != nil {
			return base.ErrorWithMessage(err, "failed building S3 client: %s", err.Error())
		}

		loc, err := c.GetBucketLocation(ctx, s.Bucket)
		if err != nil {
			return base.ErrorWithMessage(err, "failed to determine bucket location/region")
		}

		err = s.UpdateDeterminedRegion(q, &loc)
		if err != nil {
			return base.InternalError(err)
		}
	}

	return base.ReconcileResult{}
}
