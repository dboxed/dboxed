package volume_providers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/reconcilers/volume_providers/forget"
	"github.com/dboxed/dboxed/pkg/reconcilers/volume_providers/rustic"
	"github.com/dboxed/dboxed/pkg/server/config"
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

func (r *reconciler) GetItem(ctx context.Context, id string) (*dmodel.VolumeProvider, error) {
	return dmodel.GetVolumeProviderById(querier.GetQuerier(ctx), nil, id, false)
}

func (r *reconciler) getSubReconciler(mp *dmodel.VolumeProvider) (subReconciler, error) {
	switch mp.Type {
	case dmodel.VolumeProviderTypeRustic:
		return &rustic.Reconciler{}, nil
	default:
		return nil, fmt.Errorf("unsupported volume provider type %s", mp.Type)
	}
}

func (r *reconciler) getVolumeProviderChildren(ctx context.Context, vp *dmodel.VolumeProvider) (map[string]*dmodel.VolumeWithJoins, map[string]*dmodel.VolumeSnapshot, base.ReconcileResult) {
	q := querier.GetQuerier(ctx)
	volumes, err := dmodel.ListVolumesForVolumeProvider(q, vp.ID, false)
	if err != nil {
		return nil, nil, base.InternalError(err)
	}
	volumesById := map[string]*dmodel.VolumeWithJoins{}
	for _, v := range volumes {
		volumesById[v.ID] = &v
	}

	snapshots, err := dmodel.ListVolumeSnapshotsForProvider(q, nil, vp.ID, false)
	if err != nil {
		return nil, nil, base.InternalError(err)
	}
	snapshotsById := map[string]*dmodel.VolumeSnapshot{}
	for _, s := range snapshots {
		snapshotsById[s.ID] = &s
	}
	return volumesById, snapshotsById, base.ReconcileResult{}
}

func (r *reconciler) Reconcile(ctx context.Context, vp *dmodel.VolumeProvider, log *slog.Logger) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	log = log.With(
		slog.Any("name", vp.Name),
		slog.Any("volumeProviderType", vp.Type),
	)

	sr, err := r.getSubReconciler(vp)
	if err != nil {
		return base.InternalError(err)
	}

	if vp.GetDeletedAt() != nil {
		return base.ReconcileResult{}
	}

	dbVolumes, dbSnapshots, result := r.getVolumeProviderChildren(ctx, vp)
	if result.Error != nil {
		return result
	}

	result = r.forgetOldSnapshots(ctx, log, vp, dbVolumes, dbSnapshots)
	if result.Error != nil {
		return result
	}

	result = sr.ReconcileVolumeProvider(ctx, log, vp, dbVolumes, dbSnapshots)
	if result.Error != nil {
		return result
	}

	for _, s := range dbSnapshots {
		if s.DeletedAt.Valid {
			v, ok := dbVolumes[s.VolumedID.V]
			if !ok {
				return base.ErrorFromMessage("volume %s for snapshot %s not found", s.VolumedID.V, s.ID)
			}

			result = sr.ReconcileDeleteSnapshot(ctx, log, vp, &v.Volume, s)
			if result.Error != nil {
				return result
			}

			log.InfoContext(ctx, "finally deleting snapshot", slog.Any("snapshotId", s.ID), slog.Any("rsSnapshotId", s.Rustic.SnapshotId.V))
			err = querier.DeleteOneByStruct(q, s)
			if err != nil {
				return base.InternalError(err)
			}
		}
	}

	for _, v := range dbVolumes {
		if v.DeletedAt.Valid {
			result = sr.ReconcileDeleteVolume(ctx, log, vp, v)
			if result.Error != nil {
				return result
			}

			log.InfoContext(ctx, "finally deleting volume", slog.Any("volumeId", v.ID))
			err = querier.DeleteOneByStruct(q, v)
			if err != nil {
				return base.InternalError(err)
			}
		}
	}

	return base.ReconcileResult{}
}

func (r *reconciler) forgetOldSnapshots(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, dbVolumes map[string]*dmodel.VolumeWithJoins, dbSnapshots map[string]*dmodel.VolumeSnapshot) base.ReconcileResult {
	snapshotsByVolumes := map[string][]*dmodel.VolumeSnapshot{}

	for _, s := range dbSnapshots {
		if s.DeletedAt.Valid {
			continue
		}
		snapshotsByVolumes[s.VolumedID.V] = append(snapshotsByVolumes[s.VolumedID.V], s)
	}

	for _, v := range dbVolumes {
		sl, ok := snapshotsByVolumes[v.ID]
		if !ok {
			continue
		}
		result := r.forgetOldSnapshotsForVolume(ctx, log, v, sl)
		if result.Error != nil {
			return result
		}
	}

	return base.ReconcileResult{}
}

func (r *reconciler) forgetOldSnapshotsForVolume(ctx context.Context, log *slog.Logger, v *dmodel.VolumeWithJoins, snapshots []*dmodel.VolumeSnapshot) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	p := forget.ExpirePolicy{
		Last:    2,
		Hourly:  6,
		Daily:   7,
		Weekly:  4,
		Monthly: 6,
		Yearly:  1,
	}

	_, remove, _ := forget.ApplyPolicy(snapshots, p)
	for _, s := range remove {
		log.InfoContext(ctx, "marking old snapshot for deletion", slog.Any("volumeId", v.ID), slog.Any("snapshotId", s.ID))
		err := dmodel.SoftDeleteByStruct(q, s)
		if err != nil {
			return base.InternalError(err)
		}
	}
	return base.ReconcileResult{}
}

type subReconciler interface {
	ReconcileVolumeProvider(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, dbVolumes map[string]*dmodel.VolumeWithJoins, dbSnapshots map[string]*dmodel.VolumeSnapshot) base.ReconcileResult
	ReconcileDeleteSnapshot(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, dbVolume *dmodel.Volume, dbSnapshot *dmodel.VolumeSnapshot) base.ReconcileResult
	ReconcileDeleteVolume(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, dbVolume *dmodel.VolumeWithJoins) base.ReconcileResult
}
