package machines

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

func NewMachinesReconciler(config config.Config) *base.Reconciler[*dmodel.Machine] {
	return base.NewReconciler(base.Config[*dmodel.Machine]{
		ServerConfig:   config,
		ReconcilerName: "machines",
		Impl:           &reconciler{},
	})
}

func (r *reconciler) GetItem(ctx context.Context, id string) (*dmodel.Machine, error) {
	return dmodel.GetMachineById(querier.GetQuerier(ctx), nil, id, false)
}

func (r *reconciler) Reconcile(ctx context.Context, m *dmodel.Machine, log *slog.Logger) base.ReconcileResult {
	log = slog.With(
		slog.Any("name", m.Name),
	)

	return base.ReconcileResult{}
}
