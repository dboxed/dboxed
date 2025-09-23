package networks

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/reconcilers/networks/netbird"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dbutils"
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

func (r *networkReconciler) GetItem(ctx context.Context, id int64) (*dmodel.Network, error) {
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

func (r *networkReconciler) Reconcile(ctx context.Context, n *dmodel.Network, log *slog.Logger) error {
	q := querier.GetQuerier(ctx)

	log = log.With(
		slog.Any("networkName", n.Name),
		slog.Any("networkType", n.Type),
	)

	sr, err := r.getSubReconciler(ctx, n)
	if err != nil {
		return err
	}

	err = sr.Reconcile(ctx, log, n)
	if err != nil {
		return err
	}

	err = dbutils.DoAndFindChanged(ctx, func() ([]dmodel.Box, error) {
		return dmodel.ListBoxesForNetwork(q, n.ID, false)
	}, func(v dmodel.Box) error {
		return sr.ReconcileBox(ctx, &v)
	})
	if err != nil {
		return err
	}

	err = sr.Cleanup(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *networkReconciler) ReconcileDelete(ctx context.Context, n *dmodel.Network, log *slog.Logger) error {
	log = log.With(
		slog.Any("networkName", n.Name),
		slog.Any("networkType", n.Type),
	)

	sr, err := r.getSubReconciler(ctx, n)
	if err != nil {
		return err
	}

	err = sr.ReconcileDelete(ctx, log, n)
	if err != nil {
		return err
	}
	return nil
}

type subReconciler interface {
	Reconcile(ctx context.Context, log *slog.Logger, n *dmodel.Network) error
	ReconcileDelete(ctx context.Context, log *slog.Logger, n *dmodel.Network) error
	Cleanup(ctx context.Context) error
	ReconcileBox(ctx context.Context, box *dmodel.Box) error
}
