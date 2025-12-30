package boxes

import (
	"context"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
)

type reconciler struct {
}

func NewBoxesReconciler() *base.Reconciler[*dmodel.BoxWithSandbox] {
	return base.NewReconciler(base.Config[*dmodel.BoxWithSandbox]{
		ReconcilerName:        "boxes",
		FullReconcileInterval: time.Second * 60,
		Reconciler:            &reconciler{},
	})
}

func (r *reconciler) GetItem(ctx context.Context, id string) (*dmodel.BoxWithSandbox, error) {
	return dmodel.GetBoxWithSandboxById(querier.GetQuerier(ctx), nil, id, false)
}

func (r *reconciler) Reconcile(ctx context.Context, box *dmodel.BoxWithSandbox, log *slog.Logger) base.ReconcileResult {
	log = log.With(
		slog.Any("name", box.Name),
	)

	if box.Sandbox == nil || !box.Sandbox.ID.Valid {
		return base.StatusWithMessage("New", "Box is new and has no sandbox status yet")
	}

	// Check if status is stale (older than 60 seconds)
	if box.Sandbox.StatusTime != nil {
		if box.Sandbox.RunStatus != nil && *box.Sandbox.RunStatus == "stopped" {
			return base.StatusWithMessage("Stopped", "Box stopped")
		}

		statusAge := time.Since(*box.Sandbox.StatusTime)
		if statusAge > 60*time.Second {
			if !box.Enabled {
				return base.StatusWithMessage("Disabled", "Box disabled")
			}
			return base.StatusWithMessage("Stale", "Sandbox status is stale")
		}
		if box.Sandbox.RunStatus != nil {
			return base.StatusWithMessage(*box.Sandbox.RunStatus, "")
		}
	}

	return base.ReconcileResult{}
}
