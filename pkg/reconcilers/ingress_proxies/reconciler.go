package ingress_proxies

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
	log = log.With(
		slog.Any("name", proxy.Name),
	)

	if proxy.GetDeletedAt() != nil {
		return r.reconcileDelete(ctx, proxy, log)
	}

	result := r.reconcileBoxReplicas(ctx, proxy, log)
	if result.Error != nil {
		return result
	}

	return base.ReconcileResult{}
}

func (r *reconciler) reconcileDelete(ctx context.Context, proxy *dmodel.IngressProxy, log *slog.Logger) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	existingReplicas, err := dmodel.ListIngressProxyBoxesForProxy(q, proxy.ID)
	if err != nil {
		return base.InternalError(err)
	}

	for _, replica := range existingReplicas {
		result := r.softDeleteReplica(ctx, proxy, replica.BoxId, log)
		if result.Error != nil {
			return result
		}
	}

	return base.ReconcileResult{}
}

func (r *reconciler) softDeleteReplica(ctx context.Context, proxy *dmodel.IngressProxy, boxId string, log *slog.Logger) base.ReconcileResult {
	return base.Transaction(ctx, func(ctx context.Context) base.ReconcileResult {
		q := querier.GetQuerier(ctx)
		box, err := dmodel.GetBoxById(q, nil, boxId, false)
		if err != nil {
			return base.InternalError(err)
		}

		err = querier.DeleteOneByFields[dmodel.IngressProxyBox](q, map[string]any{
			"ingress_proxy_id": proxy.ID,
			"box_id":           box.ID,
		})
		if err != nil {
			return base.InternalError(err)
		}

		if box.GetDeletedAt() == nil {
			log.InfoContext(ctx, "cascading delete to ingress proxy box", "boxId", box.ID, "boxName", box.Name)
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
	})
}
