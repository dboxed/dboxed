package workspaces

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

func NewWorkspacesReconciler(config config.Config) *base.Reconciler[*dmodel.Workspace] {
	return base.NewReconciler(base.Config[*dmodel.Workspace]{
		ServerConfig:          config,
		ReconcilerName:        "workspaces",
		Impl:                  &reconciler{},
		FullReconcileInterval: 10 * time.Second,
	})
}

func (r *reconciler) GetItem(ctx context.Context, id int64) (*dmodel.Workspace, error) {
	return dmodel.GetWorkspaceById(querier.GetQuerier(ctx), id, false)
}

func (r *reconciler) Reconcile(ctx context.Context, w *dmodel.Workspace, log *slog.Logger) base.ReconcileResult {
	log = slog.With(
		slog.Any("name", w.Name),
	)

	result := r.reconcileLogQuotas(ctx, w, log)
	if result.Error != nil {
		return result
	}

	return base.ReconcileResult{}
}
