package networks

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/reconcilers/networks/netbird"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
)

type networkReconciler struct {
}

func NewNetworksReconciler(config config.Config) *base.Reconciler[*dmodel.Network] {
	return base.NewReconciler(base.Config[*dmodel.Network]{
		ServerConfig:          config,
		ReconcilerName:        "networks",
		Impl:                  &networkReconciler{},
		FullReconcileInterval: 5 * time.Second,
	})
}

func (r *networkReconciler) GetItem(ctx context.Context, id string) (*dmodel.Network, error) {
	return dmodel.GetNetworkById(querier.GetQuerier(ctx), nil, id, false)
}

func (r *networkReconciler) getSubReconciler(ctx context.Context, n *dmodel.Network) (subReconciler, error) {
	switch global.NetworkType(n.Type) {
	case global.NetworkNetbird:
		return &netbird.Reconciler{}, nil
	default:
		return nil, fmt.Errorf("unsupported network type %s", n.Type)
	}
}

func (r *networkReconciler) Reconcile(ctx context.Context, n *dmodel.Network, log *slog.Logger) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	log = log.With(
		slog.Any("networkName", n.Name),
		slog.Any("networkType", n.Type),
	)

	sr, err := r.getSubReconciler(ctx, n)
	if err != nil {
		return base.InternalError(err)
	}

	result := sr.ReconcileNetwork(ctx, log, n)
	if result.Error != nil {
		return result
	}

	if n.GetDeletedAt() != nil {
		return base.ReconcileResult{}
	}

	boxes, err := dmodel.ListBoxesForNetwork(q, n.ID, false)
	if err != nil {
		return base.InternalError(err)
	}

	for _, box := range boxes {
		sr.ReconcileBox(ctx, log, &box.Box)
	}

	result = sr.Cleanup(ctx)
	if result.Error != nil {
		return result
	}

	return result
}

type subReconciler interface {
	ReconcileNetwork(ctx context.Context, log *slog.Logger, n *dmodel.Network) base.ReconcileResult
	Cleanup(ctx context.Context) base.ReconcileResult
	ReconcileBox(ctx context.Context, log *slog.Logger, box *dmodel.Box)
}
