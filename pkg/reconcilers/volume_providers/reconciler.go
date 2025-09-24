package volume_providers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/reconcilers/volume_providers/rustic"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dbutils"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
)

type reconciler struct {
}

func NewVolumeProvidersReconciler(config config.Config) *base.Reconciler[*dmodel.VolumeProvider] {
	return base.NewReconciler(base.Config[*dmodel.VolumeProvider]{
		ServerConfig:          config,
		ReconcilerName:        "volume_providers",
		FullReconcileInterval: 5 * time.Second,
		Impl:                  &reconciler{},
	})
}

func (r *reconciler) GetItem(ctx context.Context, id int64) (*dmodel.VolumeProvider, error) {
	return dmodel.GetVolumeProviderById(querier.GetQuerier(ctx), nil, id, false)
}

func (r *reconciler) getSubReconciler(mp *dmodel.VolumeProvider) (subReconciler, error) {
	switch dmodel.VolumeProviderType(mp.Type) {
	case dmodel.VolumeProviderTypeRustic:
		return &rustic.Reconciler{}, nil
	default:
		return nil, fmt.Errorf("unsupported volume provider type %s", mp.Type)
	}
}

func (r *reconciler) Reconcile(ctx context.Context, vp *dmodel.VolumeProvider, log *slog.Logger) error {
	q := querier.GetQuerier(ctx)

	log = log.With(
		slog.Any("name", vp.Name),
		slog.Any("volumeProviderType", vp.Type),
	)

	sr, err := r.getSubReconciler(vp)
	if err != nil {
		return err
	}

	err = sr.ReconcileVolumeProvider(ctx, log, vp)
	if err != nil {
		return err
	}

	err = dbutils.DoAndFindChanged(ctx, func() ([]dmodel.VolumeWithAttachment, error) {
		return dmodel.ListVolumesForVolumeProvider(q, vp.ID, false)
	}, func(v dmodel.VolumeWithAttachment) error {
		return sr.ReconcileVolume(ctx, &v.Volume)
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *reconciler) ReconcileDelete(ctx context.Context, vp *dmodel.VolumeProvider, log *slog.Logger) error {
	log = log.With(
		slog.Any("name", vp.Name),
		slog.Any("volumeProviderType", vp.Type),
	)

	sr, err := r.getSubReconciler(vp)
	if err != nil {
		return err
	}

	err = sr.ReconcileDeleteVolumeProvider(ctx, log, vp)
	if err != nil {
		return err
	}
	return nil
}

type subReconciler interface {
	ReconcileVolumeProvider(ctx context.Context, log *slog.Logger, mp *dmodel.VolumeProvider) error
	ReconcileDeleteVolumeProvider(ctx context.Context, log *slog.Logger, mp *dmodel.VolumeProvider) error
	ReconcileVolume(ctx context.Context, m *dmodel.Volume) error
}
