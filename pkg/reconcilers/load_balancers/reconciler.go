package load_balancers

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

func NewLoadBalancersReconciler(config config.Config) *base.Reconciler[*dmodel.LoadBalancer] {
	return base.NewReconciler(base.Config[*dmodel.LoadBalancer]{
		ServerConfig:          config,
		ReconcilerName:        "load-balancers",
		FullReconcileInterval: time.Second * 10,
		Impl:                  &reconciler{},
	})
}

func (r *reconciler) GetItem(ctx context.Context, id string) (*dmodel.LoadBalancer, error) {
	return dmodel.GetLoadBalancerById(querier.GetQuerier(ctx), nil, id, false)
}

func (r *reconciler) Reconcile(ctx context.Context, lb *dmodel.LoadBalancer, log *slog.Logger) base.ReconcileResult {
	log = log.With(
		slog.Any("name", lb.Name),
	)

	if lb.GetDeletedAt() != nil {
		return r.reconcileDelete(ctx, lb, log)
	}

	result := r.reconcileBoxReplicas(ctx, lb, log)
	if result.Error != nil {
		return result
	}

	return base.ReconcileResult{}
}

func (r *reconciler) reconcileDelete(ctx context.Context, lb *dmodel.LoadBalancer, log *slog.Logger) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	existingReplicas, err := dmodel.ListLoadBalancerBoxesForLoadBalancer(q, lb.ID)
	if err != nil {
		return base.InternalError(err)
	}

	for _, replica := range existingReplicas {
		result := r.softDeleteReplica(ctx, lb, replica.BoxId, log)
		if result.Error != nil {
			return result
		}
	}

	return base.ReconcileResult{}
}

func (r *reconciler) softDeleteReplica(ctx context.Context, lb *dmodel.LoadBalancer, boxId string, log *slog.Logger) base.ReconcileResult {
	return base.Transaction(ctx, func(ctx context.Context) base.ReconcileResult {
		q := querier.GetQuerier(ctx)
		box, err := dmodel.GetBoxById(q, nil, boxId, false)
		if err != nil {
			return base.InternalError(err)
		}

		err = querier.DeleteOneByFields[dmodel.LoadBalancerBox](q, map[string]any{
			"load_balancer_id": lb.ID,
			"box_id":           box.ID,
		})
		if err != nil {
			return base.InternalError(err)
		}

		if box.GetDeletedAt() == nil {
			log.InfoContext(ctx, "cascading delete to load-balancer box", "boxId", box.ID, "boxName", box.Name)
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
