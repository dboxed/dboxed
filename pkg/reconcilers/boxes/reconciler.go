package boxes

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

func NewBoxesReconciler(config config.Config) *base.Reconciler[*dmodel.Box] {
	return base.NewReconciler(base.Config[*dmodel.Box]{
		ServerConfig:   config,
		ReconcilerName: "boxes",
		Impl:           &reconciler{},
	})
}

func (r *reconciler) GetItem(ctx context.Context, id int64) (*dmodel.Box, error) {
	return dmodel.GetBoxById(querier.GetQuerier(ctx), nil, id, false)
}

func (r *reconciler) Reconcile(ctx context.Context, box *dmodel.Box, log *slog.Logger) base.ReconcileResult {
	log = log.With(
		slog.Any("name", box.Name),
	)

	return base.ReconcileResult{}
}
