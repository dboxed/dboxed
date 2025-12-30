package load_balancers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
)

type reconciler struct {
}

func NewLoadBalancersReconciler() *base.Reconciler[*dmodel.LoadBalancer] {
	return base.NewReconciler(base.Config[*dmodel.LoadBalancer]{
		ReconcilerName:        "load-balancers",
		FullReconcileInterval: time.Second * 10,
		Reconciler:            &reconciler{},
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

	token, result := r.reconcileToken(ctx, lb, log)
	if result.ExitReconcile() {
		return result
	}

	result = r.reconcileBoxReplicas(ctx, lb, token, log)
	if result.ExitReconcile() {
		return result
	}

	// Calculate and set status based on box replicas
	return r.calculateStatus(ctx, lb, log)
}

func (r *reconciler) reconcileDelete(ctx context.Context, lb *dmodel.LoadBalancer, log *slog.Logger) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	existingReplicas, err := dmodel.ListLoadBalancerBoxesForLoadBalancer(q, lb.ID)
	if err != nil {
		return base.InternalError(err)
	}

	for _, replica := range existingReplicas {
		result := r.softDeleteReplica(ctx, lb, replica.BoxId, log)
		if result.ExitReconcile() {
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
			err = dmodel.BumpChangeSeq(q, box)
			if err != nil {
				return base.InternalError(err)
			}
		}

		return base.ReconcileResult{}
	})
}

func (r *reconciler) calculateStatus(ctx context.Context, lb *dmodel.LoadBalancer, log *slog.Logger) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	// Get all box replicas for this load balancer
	replicas, err := dmodel.ListLoadBalancerBoxesForLoadBalancer(q, lb.ID)
	if err != nil {
		return base.InternalError(err)
	}

	if len(replicas) == 0 {
		return base.StatusWithMessage("NoReplicas", "Load balancer has no box replicas")
	}

	// Track box statuses
	totalBoxes := len(replicas)
	readyBoxes := 0
	staleBoxes := 0
	errorBoxes := 0
	deletedBoxes := 0

	for _, replica := range replicas {
		// Get box details
		box, err := dmodel.GetBoxWithSandboxById(q, nil, replica.BoxId, false)
		if err != nil {
			if querier.IsSqlNotFoundError(err) {
				deletedBoxes++
				continue
			}
			return base.InternalError(err)
		}

		// Skip deleted boxes
		if box.GetDeletedAt() != nil {
			deletedBoxes++
			continue
		}

		// Check box reconcile status
		if box.ReconcileStatus.ReconcileStatus.V == "Error" {
			errorBoxes++
			continue
		}

		if box.Sandbox != nil || !box.Sandbox.ID.Valid {
			// Check if status is stale
			if box.Sandbox.StatusTime != nil {
				statusAge := time.Since(*box.Sandbox.StatusTime)
				if statusAge > 60*time.Second {
					staleBoxes++
					continue
				}
			}

			// Check if box is running
			if box.Sandbox.RunStatus != nil && *box.Sandbox.RunStatus == "running" {
				readyBoxes++
			}
		}
	}

	// Calculate overall status
	activeBoxes := totalBoxes - deletedBoxes

	if activeBoxes == 0 {
		return base.StatusWithMessage("NoActiveReplicas", "All box replicas have been deleted")
	}

	if errorBoxes > 0 {
		return base.StatusWithMessage("Degraded", fmt.Sprintf("%d/%d replicas in error state", errorBoxes, activeBoxes))
	}

	if readyBoxes == 0 {
		return base.StatusWithMessage("Unavailable", "No replicas are running")
	}

	if staleBoxes > 0 {
		return base.StatusWithMessage("Degraded", fmt.Sprintf("%d/%d replicas ready, %d stale", readyBoxes, activeBoxes, staleBoxes))
	}

	if readyBoxes < activeBoxes {
		return base.StatusWithMessage("Degraded", fmt.Sprintf("%d/%d replicas ready", readyBoxes, activeBoxes))
	}

	// All replicas are ready
	return base.StatusWithMessage("Ready", fmt.Sprintf("All %d replicas ready", readyBoxes))
}
