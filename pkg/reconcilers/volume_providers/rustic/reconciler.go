package rustic

import (
	"context"
	"errors"
	"log/slog"
	"path"
	"strings"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/s3utils"
	"github.com/minio/minio-go/v7"
)

type Reconciler struct {
}

func (r *Reconciler) listRusticSnapshotIds(ctx context.Context, vp *dmodel.VolumeProvider) ([]string, base.ReconcileResult) {
	b, c, err := s3utils.BuildS3ClientFromId(ctx, *vp.Rustic.S3BucketID, "")
	if err != nil {
		return nil, base.ErrorWithMessage(err, "failed building S3 client: %s", err.Error())
	}

	prefix := path.Join(vp.Rustic.StoragePrefix.V, "snapshots") + "/"
	ch := c.ListObjects(ctx, b.Bucket, minio.ListObjectsOptions{
		Prefix: prefix,
	})
	defer func() {
		// drain it
		for range ch {
		}
	}()

	var ret []string
	for oi := range ch {
		if oi.Err != nil {
			return nil, base.ErrorWithMessage(oi.Err, "failed to list S3 objects for %s: %s", prefix, oi.Err.Error())
		}
		id := path.Base(oi.Key)
		ret = append(ret, id)
	}

	return ret, base.ReconcileResult{}
}

func (r *Reconciler) deleteRusticSnapshot(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, snapshotId string) base.ReconcileResult {
	log.InfoContext(ctx, "deleting rustic snapshot", slog.Any("rsSnapshotId", snapshotId))

	b, c, err := s3utils.BuildS3ClientFromId(ctx, *vp.Rustic.S3BucketID, "")
	if err != nil {
		return base.ErrorWithMessage(err, "failed building S3 client: %s", err.Error())
	}

	prefix := path.Join(vp.Rustic.StoragePrefix.V, "snapshots") + "/"

	key := path.Join(prefix, snapshotId)
	err = c.RemoveObject(ctx, b.Bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		var err2 *minio.ErrorResponse
		if errors.As(err, &err2) {
			if err2.Code == minio.NoSuchKey {
				return base.ReconcileResult{}
			}
		}
		return base.ErrorWithMessage(err, "failed to remove snapshot via S3 RemoveObject: %s", err.Error())
	}
	return base.ReconcileResult{}
}

func (r *Reconciler) getTagValue(tags []string, tagPrefix string) string {
	tagPrefix += "-"
	for _, tag := range tags {
		if strings.HasPrefix(tag, tagPrefix) {
			return strings.TrimPrefix(tag, tagPrefix)
		}
	}
	return ""
}

func (r *Reconciler) ReconcileVolumeProvider(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, dbVolumes map[int64]*dmodel.VolumeWithAttachment, dbSnapshots map[int64]*dmodel.VolumeSnapshot) base.ReconcileResult {
	dbSnapshotsByRusticId := map[string]*dmodel.VolumeSnapshot{}
	for _, s := range dbSnapshots {
		dbSnapshotsByRusticId[s.Rustic.SnapshotId.V] = s
	}
	dbVolumesByUuuid := map[string]*dmodel.VolumeWithAttachment{}
	for _, v := range dbVolumes {
		dbVolumesByUuuid[v.Uuid] = v
	}

	rsSnapshotIds, result := r.listRusticSnapshotIds(ctx, vp)
	if result.Error != nil {
		return result
	}
	rsSnapshotIdSet := map[string]struct{}{}
	for _, id := range rsSnapshotIds {
		rsSnapshotIdSet[id] = struct{}{}
	}

	for _, s := range dbSnapshots {
		if s.DeletedAt.Valid {
			continue
		}
		log := log.With(
			slog.Any("volumeId", s.VolumedID.V),
			slog.Any("snapshotId", s.ID),
			slog.Any("snapshotRusticId", s.Rustic.SnapshotId.V),
		)

		_, ok := rsSnapshotIdSet[s.Rustic.SnapshotId.V]
		if !ok {
			log.InfoContext(ctx, "snapshot vanished from rustic, marking for deletion in DB")
			err := querier.Transaction(ctx, func(ctx context.Context) error {
				q := querier.GetQuerier(ctx)
				v, ok := dbVolumes[s.VolumedID.V]
				if ok {
					if v.LatestSnapshotId != nil && *v.LatestSnapshotId == s.ID {
						log.InfoContext(ctx, "snapshot was the latest snapshot, resetting latest snapshot field")
						err := v.UpdateLatestSnapshot(q, nil)
						if err != nil {
							return err
						}
					}
				}
				err := dmodel.SoftDeleteByStruct(q, s)
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return base.InternalError(err)
			}
		}
	}

	return base.ReconcileResult{}
}

func (r *Reconciler) ReconcileDeleteSnapshot(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, dbVolume *dmodel.Volume, dbSnapshot *dmodel.VolumeSnapshot) base.ReconcileResult {
	result := r.deleteRusticSnapshot(ctx, log, vp, dbSnapshot.Rustic.SnapshotId.V)
	if result.Error != nil {
		return result
	}
	return base.ReconcileResult{}
}

func (r *Reconciler) ReconcileDeleteVolume(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, dbVolume *dmodel.VolumeWithAttachment) base.ReconcileResult {
	// we assume that all snapshots have been deleted already
	return base.ReconcileResult{}
}
