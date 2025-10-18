package rustic

import (
	"context"
	"fmt"
	"log/slog"
	"path"
	"strings"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/s3utils"
	"github.com/minio/minio-go/v7"
)

type Reconciler struct {
}

func (r *Reconciler) listRusticSnapshotIds(ctx context.Context, vp *dmodel.VolumeProvider) ([]string, error) {
	c, err := s3utils.BuildS3Client(vp, "")
	if err != nil {
		return nil, err
	}

	prefix := path.Join(vp.Rustic.StorageS3.Prefix.V, "snapshots") + "/"
	ch := c.ListObjects(ctx, vp.Rustic.StorageS3.Bucket.V, minio.ListObjectsOptions{
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
			return nil, oi.Err
		}
		id := path.Base(oi.Key)
		ret = append(ret, id)
	}

	return ret, nil
}

func (r *Reconciler) deleteRusticSnapshots(ctx context.Context, vp *dmodel.VolumeProvider, snapshotIds []string) error {
	slog.InfoContext(ctx, "deleting rustic snapshots", slog.Any("rsSnapshotIds", snapshotIds))

	c, err := s3utils.BuildS3Client(vp, "")
	if err != nil {
		return err
	}

	prefix := path.Join(vp.Rustic.StorageS3.Prefix.V, "snapshots") + "/"

	for _, id := range snapshotIds {
		key := path.Join(prefix, id)
		err = c.RemoveObject(ctx, vp.Rustic.StorageS3.Bucket.V, key, minio.RemoveObjectOptions{})
		if err != nil {
			return fmt.Errorf("failed to remove snapshot via S3 RemoveObject: %w", err)
		}
	}
	return nil
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

func (r *Reconciler) ReconcileVolumeProvider(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, dbVolumes map[int64]*dmodel.VolumeWithAttachment, dbSnapshots map[int64]*dmodel.VolumeSnapshot) error {
	q := querier.GetQuerier(ctx)

	dbSnapshotsByRusticId := map[string]*dmodel.VolumeSnapshot{}
	for _, s := range dbSnapshots {
		dbSnapshotsByRusticId[s.Rustic.SnapshotId.V] = s
	}
	dbVolumesByUuuid := map[string]*dmodel.VolumeWithAttachment{}
	for _, v := range dbVolumes {
		dbVolumesByUuuid[v.Uuid] = v
	}

	rsSnapshotIds, err := r.listRusticSnapshotIds(ctx, vp)
	if err != nil {
		return fmt.Errorf("failed to list rustic snapshot ids: %w", err)
	}
	rsSnapshotIdSet := map[string]struct{}{}
	for _, id := range rsSnapshotIds {
		rsSnapshotIdSet[id] = struct{}{}
	}

	var rsSnapshotsToDelete []string
	for _, id := range rsSnapshotIds {
		dbSnapshot, ok := dbSnapshotsByRusticId[id]
		if ok && dbSnapshot.DeletedAt.Valid {
			rsSnapshotsToDelete = append(rsSnapshotsToDelete, id)
		}
	}

	if len(rsSnapshotsToDelete) != 0 {
		err = r.deleteRusticSnapshots(ctx, vp, rsSnapshotsToDelete)
		if err != nil {
			return fmt.Errorf("failed to delete snapshots: %w", err)
		}
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
			err = dmodel.SoftDeleteByStruct(q, s)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Reconciler) ReconcileDeleteVolumeProvider(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, volumes map[int64]*dmodel.VolumeWithAttachment, snapshots map[int64]*dmodel.VolumeSnapshot) error {
	return nil
}
