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

func (r *reconciler) GetItem(ctx context.Context, id int64) (*dmodel.VolumeProvider, error) {
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

func (r *reconciler) getVolumeProviderChildren(ctx context.Context, vp *dmodel.VolumeProvider) (map[int64]*dmodel.VolumeWithAttachment, map[int64]*dmodel.VolumeSnapshot, error) {
	q := querier.GetQuerier(ctx)
	volumes, err := dmodel.ListVolumesForVolumeProvider(q, vp.ID, false)
	if err != nil {
		return nil, nil, err
	}
	volumesById := map[int64]*dmodel.VolumeWithAttachment{}
	for _, v := range volumes {
		volumesById[v.ID] = &v
	}

	snapshots, err := dmodel.ListVolumeSnapshotsForProvider(q, nil, vp.ID, false)
	if err != nil {
		return nil, nil, err
	}
	snapshotsById := map[int64]*dmodel.VolumeSnapshot{}
	for _, s := range snapshots {
		snapshotsById[s.ID] = &s
	}
	return volumesById, snapshotsById, nil
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

	dbVolumes, dbSnapshots, err := r.getVolumeProviderChildren(ctx, vp)
	if err != nil {
		return err
	}

	err = r.forgetOldSnapshots(ctx, log, vp, dbVolumes, dbSnapshots)
	if err != nil {
		return err
	}

	err = sr.ReconcileVolumeProvider(ctx, log, vp, dbVolumes, dbSnapshots)
	if err != nil {
		return err
	}

	for _, s := range dbSnapshots {
		if s.DeletedAt.Valid {
			v, ok := dbVolumes[s.VolumedID.V]
			if !ok {
				return fmt.Errorf("volume %d for snapshot %d not found", s.VolumedID.V, s.ID)
			}

			err = sr.ReconcileDeleteSnapshot(ctx, log, vp, &v.Volume, s)
			if err != nil {
				return err
			}

			log.InfoContext(ctx, "finally deleting snapshot", slog.Any("snapshotId", s.ID), slog.Any("rsSnapshotId", s.Rustic.SnapshotId.V))
			err = querier.DeleteOneByStruct(q, s)
			if err != nil {
				return err
			}
		}
	}

	for _, v := range dbVolumes {
		if v.DeletedAt.Valid {
			err = sr.ReconcileDeleteVolume(ctx, log, vp, v)
			if err != nil {
				return err
			}

			log.InfoContext(ctx, "finally deleting volume", slog.Any("volumeId", v.ID))
			err = querier.DeleteOneByStruct(q, v)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *reconciler) forgetOldSnapshots(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, dbVolumes map[int64]*dmodel.VolumeWithAttachment, dbSnapshots map[int64]*dmodel.VolumeSnapshot) error {
	snapshotsByVolumes := map[int64][]*dmodel.VolumeSnapshot{}

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
		err := r.forgetOldSnapshotsForVolume(ctx, log, v, sl)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *reconciler) forgetOldSnapshotsForVolume(ctx context.Context, log *slog.Logger, v *dmodel.VolumeWithAttachment, snapshots []*dmodel.VolumeSnapshot) error {
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
			return err
		}
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

	volumes, snapshots, err := r.getVolumeProviderChildren(ctx, vp)
	if err != nil {
		return err
	}

	err = sr.ReconcileDeleteVolumeProvider(ctx, log, vp, volumes, snapshots)
	if err != nil {
		return err
	}
	return nil
}

type subReconciler interface {
	ReconcileVolumeProvider(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, dbVolumes map[int64]*dmodel.VolumeWithAttachment, dbSnapshots map[int64]*dmodel.VolumeSnapshot) error
	ReconcileDeleteVolumeProvider(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, dbVolumes map[int64]*dmodel.VolumeWithAttachment, dbSnapshots map[int64]*dmodel.VolumeSnapshot) error
	ReconcileDeleteSnapshot(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, dbVolume *dmodel.Volume, dbSnapshot *dmodel.VolumeSnapshot) error
	ReconcileDeleteVolume(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, dbVolume *dmodel.VolumeWithAttachment) error
}
