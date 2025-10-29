package machine_providers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/reconcilers/machine_providers/aws"
	"github.com/dboxed/dboxed/pkg/reconcilers/machine_providers/hetzner"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
)

type reconciler struct {
}

func NewMachineProvidersReconciler(config config.Config) *base.Reconciler[*dmodel.MachineProvider] {
	return base.NewReconciler(base.Config[*dmodel.MachineProvider]{
		ServerConfig:          config,
		ReconcilerName:        "machine_providers",
		FullReconcileInterval: 5 * time.Second,
		Impl:                  &reconciler{},
	})
}

func (r *reconciler) GetItem(ctx context.Context, id int64) (*dmodel.MachineProvider, error) {
	return dmodel.GetMachineProviderById(querier.GetQuerier(ctx), nil, id, false)
}

func (r *reconciler) getSubReconciler(mp *dmodel.MachineProvider) (subReconciler, error) {
	switch global.MachineProviderType(mp.Type) {
	case global.MachineProviderAws:
		return &aws.Reconciler{}, nil
	case global.MachineProviderHetzner:
		return &hetzner.Reconciler{}, nil
	default:
		return nil, fmt.Errorf("unsupported machine provider type %s", mp.Type)
	}
}

func (r *reconciler) Reconcile(ctx context.Context, mp *dmodel.MachineProvider, log *slog.Logger) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	log = log.With(
		slog.Any("name", mp.Name),
		slog.Any("machineProviderType", mp.Type),
	)

	sr, err := r.getSubReconciler(mp)
	if err != nil {
		return base.ReconcileResult{Error: err}
	}

	result := sr.ReconcileMachineProvider(ctx, log, mp)
	if result.Error != nil {
		return result
	}

	machines, err := dmodel.ListMachinesForMachineProvider(q, mp.ID, false)
	if err != nil {
		return base.InternalError(err)
	}

	for _, m := range machines {
		sr.ReconcileMachine(ctx, log, &m)
	}

	return base.ReconcileResult{}
}

type subReconciler interface {
	ReconcileMachineProvider(ctx context.Context, log *slog.Logger, mp *dmodel.MachineProvider) base.ReconcileResult
	ReconcileMachine(ctx context.Context, log *slog.Logger, m *dmodel.Machine)
}
