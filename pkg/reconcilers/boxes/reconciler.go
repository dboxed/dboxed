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

func NewBoxesReconciler() *base.Reconciler[*dmodel.BoxWithSandboxStatus] {
	return base.NewReconciler(base.Config[*dmodel.BoxWithSandboxStatus]{
		ReconcilerName:        "boxes",
		FullReconcileInterval: time.Second * 60,
		Reconciler:            &reconciler{},
	})
}

func (r *reconciler) GetItem(ctx context.Context, id string) (*dmodel.BoxWithSandboxStatus, error) {
	return dmodel.GetBoxWithSandboxStatusById(querier.GetQuerier(ctx), nil, id, false)
}

func (r *reconciler) Reconcile(ctx context.Context, box *dmodel.BoxWithSandboxStatus, log *slog.Logger) base.ReconcileResult {
	log = log.With(
		slog.Any("name", box.Name),
	)

	// Check if status is stale (older than 60 seconds)
	if box.SandboxStatus.StatusTime != nil {
		if box.SandboxStatus.RunStatus != nil && *box.SandboxStatus.RunStatus == "stopped" {
			return base.StatusWithMessage("Stopped", "Box stopped")
		}

		statusAge := time.Since(*box.SandboxStatus.StatusTime)
		if statusAge > 60*time.Second {
			if !box.Enabled {
				return base.StatusWithMessage("Disabled", "Box disabled")
			}
			return base.StatusWithMessage("Stale", "Sandbox status is stale")
		}
		if box.SandboxStatus.RunStatus != nil {
			return base.StatusWithMessage(*box.SandboxStatus.RunStatus, "")
		}
	} else {
		// No sandbox status yet, this is normal for new boxes
		return base.StatusWithMessage("New", "Box is new and has no sandbox status yet")
	}

	return base.ReconcileResult{}
}
