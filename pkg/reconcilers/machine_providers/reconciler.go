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
	"github.com/dboxed/dboxed/pkg/server/db/dbutils"
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

func (r *reconciler) Reconcile(ctx context.Context, mp *dmodel.MachineProvider, log *slog.Logger) error {
	q := querier.GetQuerier(ctx)

	log = log.With(
		slog.Any("name", mp.Name),
		slog.Any("machineProviderType", mp.Type),
	)

	sr, err := r.getSubReconciler(mp)
	if err != nil {
		return err
	}

	err = sr.ReconcileMachineProvider(ctx, log, mp)
	if err != nil {
		return err
	}

	err = dbutils.DoAndFindChanged(ctx, func() ([]dmodel.Machine, error) {
		return dmodel.ListMachinesForMachineProvider(q, mp.ID, false)
	}, func(v dmodel.Machine) error {
		return sr.ReconcileMachine(ctx, &v)
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *reconciler) ReconcileDelete(ctx context.Context, mp *dmodel.MachineProvider, log *slog.Logger) error {
	log = log.With(
		slog.Any("name", mp.Name),
		slog.Any("machineProviderType", mp.Type),
	)

	sr, err := r.getSubReconciler(mp)
	if err != nil {
		return err
	}

	err = sr.ReconcileDeleteMachineProvider(ctx, log, mp)
	if err != nil {
		return err
	}
	return nil
}

type subReconciler interface {
	ReconcileMachineProvider(ctx context.Context, log *slog.Logger, mp *dmodel.MachineProvider) error
	ReconcileDeleteMachineProvider(ctx context.Context, log *slog.Logger, mp *dmodel.MachineProvider) error
	ReconcileMachine(ctx context.Context, m *dmodel.Machine) error
}
