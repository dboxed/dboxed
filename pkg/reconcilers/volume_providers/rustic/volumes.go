package rustic

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
)

func (r *Reconciler) ReconcileVolume(ctx context.Context, v *dmodel.Volume) error {
	q := querier.GetQuerier(ctx)

	log := r.log

	if v.DeletedAt.Valid {
		return r.reconcileDeleteVolume(ctx, log, v)
	}

	err := dmodel.AddFinalizer(q, v, "rustic")
	if err != nil {
		return err
	}

	err = r.updateRusticVolumeStatus(ctx, log, v)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) reconcileDeleteVolume(ctx context.Context, log *slog.Logger, v *dmodel.Volume) error {
	q := querier.GetQuerier(ctx)

	err := dmodel.RemoveFinalizer(q, v, "rustic")
	if err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) updateRusticVolumeStatus(ctx context.Context, log *slog.Logger, v *dmodel.Volume) error {
	return nil
}
