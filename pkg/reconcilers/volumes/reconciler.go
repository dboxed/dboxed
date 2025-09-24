package volumes

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
)

type reconciler struct {
}

func NewVolumesReconciler(config config.Config) *base.Reconciler[*dmodel.VolumeWithAttachment] {
	return base.NewReconciler(base.Config[*dmodel.VolumeWithAttachment]{
		ServerConfig:   config,
		ReconcilerName: "volumes",
		Impl:           &reconciler{},
	})
}

func (r *reconciler) GetItem(ctx context.Context, id int64) (*dmodel.VolumeWithAttachment, error) {
	return dmodel.GetVolumeById(querier.GetQuerier(ctx), nil, id, false)
}

func (r *reconciler) Reconcile(ctx context.Context, v *dmodel.VolumeWithAttachment, log *slog.Logger) error {
	log = slog.With(
		slog.Any("name", v.Name),
	)

	return nil
}

func (r *reconciler) ReconcileDelete(ctx context.Context, v *dmodel.VolumeWithAttachment, log *slog.Logger) error {
	return nil
}
