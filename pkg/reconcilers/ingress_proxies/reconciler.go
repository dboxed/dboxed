package ingress_proxies

import (
	"context"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
)

type reconciler struct {
}

func NewIngressProxiesReconciler(config config.Config) *base.Reconciler[*dmodel.IngressProxy] {
	return base.NewReconciler(base.Config[*dmodel.IngressProxy]{
		ServerConfig:          config,
		ReconcilerName:        "ingress-proxies",
		FullReconcileInterval: time.Second * 10,
		Impl:                  &reconciler{},
	})
}

func (r *reconciler) GetItem(ctx context.Context, id string) (*dmodel.IngressProxy, error) {
	return dmodel.GetIngressProxyById(querier.GetQuerier(ctx), nil, id, false)
}

func (r *reconciler) Reconcile(ctx context.Context, proxy *dmodel.IngressProxy, log *slog.Logger) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	log = log.With(
		slog.Any("name", proxy.Name),
	)

	box, err := dmodel.GetBoxById(q, nil, proxy.BoxID, false)
	if err != nil {
		if !querier.IsSqlNotFoundError(err) {
			return base.InternalError(err)
		}
		// Box already deleted, nothing to do
		return base.ReconcileResult{}
	}

	if proxy.GetDeletedAt() != nil {
		return r.reconcileDelete(ctx, proxy, log, box)
	}

	result := r.reconcileBoxPortForwards(ctx, proxy, box, log)
	if result.Error != nil {
		return result
	}

	switch global.IngressProxyType(proxy.ProxyType) {
	case global.IngressProxyCaddy:
		result := r.reconcileProxyCaddy(ctx, proxy, box, log)
		if result.Error != nil {
			return result
		}
	}

	return base.ReconcileResult{}
}

func (r *reconciler) reconcileDelete(ctx context.Context, proxy *dmodel.IngressProxy, log *slog.Logger, box *dmodel.Box) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	if box.GetDeletedAt() == nil {
		log.InfoContext(ctx, "cascading delete to ingress proxy box")
		err := dmodel.SoftDeleteByStruct(q, box)
		if err != nil {
			return base.InternalError(err)
		}
		err = dmodel.AddChangeTracking(q, box)
		if err != nil {
			return base.InternalError(err)
		}
	}

	return base.ReconcileResult{}
}
