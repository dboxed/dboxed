package machines

import (
	"context"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/resources/machines"
)

type reconciler struct {
}

func NewMachinesReconciler() *base.Reconciler[*dmodel.MachineWithRunStatus] {
	return base.NewReconciler(base.Config[*dmodel.MachineWithRunStatus]{
		ReconcilerName:        "machines",
		FullReconcileInterval: time.Second * 60,
		Reconciler:            &reconciler{},
	})
}

func (r *reconciler) GetItem(ctx context.Context, id string) (*dmodel.MachineWithRunStatus, error) {
	return dmodel.GetMachineWithRunStatusById(querier.GetQuerier(ctx), nil, id, false)
}

func (r *reconciler) Reconcile(ctx context.Context, m *dmodel.MachineWithRunStatus, log *slog.Logger) base.ReconcileResult {
	log = slog.With(
		slog.Any("name", m.Name),
	)

	if m.DeletedAt.Valid {
		err := machines.InvalidateBoxTokens(ctx, m.WorkspaceID, m.ID, nil)
		if err != nil {
			return base.InternalError(err)
		}
	}

	// Check if status is stale (older than 60 seconds)
	if m.RunStatus.StatusTime != nil {
		if m.RunStatus.RunStatus != nil && *m.RunStatus.RunStatus == "stopped" {
			return base.StatusWithMessage("Stopped", "Machine stopped")
		}

		statusAge := time.Since(*m.RunStatus.StatusTime)
		if statusAge > 60*time.Second {
			return base.StatusWithMessage("Stale", "Machine run status is stale")
		}
		if m.RunStatus.RunStatus != nil {
			return base.StatusWithMessage(*m.RunStatus.RunStatus, "")
		}
	} else {
		// No run status yet, this is normal for new machines
		return base.StatusWithMessage("New", "Machine is new and has no run status yet")
	}

	return base.ReconcileResult{}
}
