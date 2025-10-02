package rustic

import (
	"context"
	"fmt"
	"log/slog"
	"path"
	"slices"
	"strings"

	"github.com/dboxed/dboxed/pkg/server/db/dbutils"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/s3utils"
	"github.com/dboxed/dboxed/pkg/volume/rustic"
	"github.com/dboxed/dboxed/pkg/volume/volume_backup"
	"github.com/minio/minio-go/v7"
)

type Reconciler struct {
}

func (r *Reconciler) buildRusticConfig(vp *dmodel.VolumeProvider) (*rustic.RusticConfig, error) {
	if vp.Rustic.StorageType != dmodel.VolumeProviderStorageTypeS3 {
		return nil, fmt.Errorf("only S3 storage is supported ")
	}
	config := &rustic.RusticConfig{
		Repository: rustic.RusticConfigRepository{
			Repository: fmt.Sprintf("opendal:s3"),
			Password:   vp.Rustic.Password.V,
			Options: rustic.RusticConfigRepositoryOptions{
				Endpoint:        vp.Rustic.StorageS3.Endpoint.V,
				Bucket:          vp.Rustic.StorageS3.Bucket.V,
				Region:          vp.Rustic.StorageS3.Region,
				AccessKeyId:     vp.Rustic.StorageS3.AccessKeyId.V,
				SecretAccessKey: vp.Rustic.StorageS3.SecretAccessKey.V,
				Root:            vp.Rustic.StorageS3.Prefix.V,
			},
		},
	}
	return config, nil
}

func (r *Reconciler) listRusticSnapshotIds(ctx context.Context, vp *dmodel.VolumeProvider) ([]string, error) {
	c, err := s3utils.BuildS3Client(vp)
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
			return nil, err
		}
		id := path.Base(oi.Key)
		ret = append(ret, id)
	}

	return ret, nil
}

func (r *Reconciler) getFilteredRusticSnapshots(ctx context.Context, vp *dmodel.VolumeProvider, snapshotIds []string) (map[string]*rustic.Snapshot, error) {
	config, err := r.buildRusticConfig(vp)
	if err != nil {
		return nil, err
	}

	snapshots, err := rustic.RunSnapshots(ctx, *config, rustic.SnapshotOpts{
		SnapshotIds: snapshotIds,
		NoCache:     true,
	})
	if err != nil {
		return nil, err
	}

	expectedTags := volume_backup.BuildBackupTags(vp.ID, nil, nil, nil)

	ret := map[string]*rustic.Snapshot{}
	for _, s := range snapshots {
		allFound := true
		for _, tag := range expectedTags {
			if !slices.Contains(s.Tags, tag) {
				allFound = false
				break
			}
		}
		if allFound {
			ret[s.Id] = &s
		}
	}

	return ret, nil
}

func (r *Reconciler) deleteRusticSnapshots(ctx context.Context, vp *dmodel.VolumeProvider, snapshotIds []string) error {
	slog.InfoContext(ctx, "deleting rustic snapshots", slog.Any("rsSnapshotIds", snapshotIds))

	c, err := s3utils.BuildS3Client(vp)
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
	var rsPotentialSnapshotsToCreate []string
	for _, id := range rsSnapshotIds {
		dbSnapshot, ok := dbSnapshotsByRusticId[id]
		if !ok {
			rsPotentialSnapshotsToCreate = append(rsPotentialSnapshotsToCreate, id)
		} else if dbSnapshot.DeletedAt.Valid {
			rsSnapshotsToDelete = append(rsSnapshotsToDelete, id)
		}
	}

	if len(rsPotentialSnapshotsToCreate) != 0 {
		rsSnapshots, err := r.getFilteredRusticSnapshots(ctx, vp, rsPotentialSnapshotsToCreate)
		if err != nil {
			return fmt.Errorf("failed to retrieve rustic snapshots: %w", err)
		}
		for _, s := range rsSnapshots {
			volumeUuid := r.getTagValue(s.Tags, "dboxed-volume-uuid")
			v := dbVolumesByUuuid[volumeUuid]

			if v == nil || v.DeletedAt.Valid {
				rsSnapshotsToDelete = append(rsSnapshotsToDelete, s.Id)
				continue
			}

			newDbSnapshot, err := r.createDBSnapshot(ctx, log, vp, &v.Volume, s)
			if err != nil {
				return fmt.Errorf("failed to create snapshot: %w", err)
			}
			dbSnapshots[newDbSnapshot.ID] = newDbSnapshot
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

func (r *Reconciler) createDBSnapshot(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, v *dmodel.Volume, rsSnapshot *rustic.Snapshot) (*dmodel.VolumeSnapshot, error) {
	var ret *dmodel.VolumeSnapshot
	err := dbutils.RunInTx(ctx, func(ctx context.Context) error {
		var err error
		ret, err = r.createDBSnapshotInTx(ctx, log, vp, v, rsSnapshot)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (r *Reconciler) createDBSnapshotInTx(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, v *dmodel.Volume, rsSnapshot *rustic.Snapshot) (*dmodel.VolumeSnapshot, error) {
	q := querier.GetQuerier(ctx)

	log.InfoContext(ctx, "creating snapshot in database", slog.Any("rsSnapshotId", rsSnapshot.Id))

	lockId := r.getTagValue(rsSnapshot.Tags, "dboxed-volume-lock-id")

	snapshot := dmodel.VolumeSnapshot{
		OwnedByWorkspace: dmodel.OwnedByWorkspace{
			WorkspaceID: vp.WorkspaceID,
		},
		VolumeProviderID: vp.ID,
		VolumedID:        querier.N(v.ID),
		LockID:           lockId,
	}

	err := snapshot.Create(q)
	if err != nil {
		return nil, err
	}

	if v.Rustic != nil {
		snapshot.Rustic = &dmodel.VolumeSnapshotRustic{
			ID:                    querier.N(snapshot.ID),
			SnapshotId:            querier.N(rsSnapshot.Id),
			SnapshotTime:          querier.N(rsSnapshot.Time),
			ParentSnapshotId:      rsSnapshot.Parent,
			Hostname:              querier.N(rsSnapshot.Hostname),
			FilesNew:              querier.N(rsSnapshot.Summary.FilesNew),
			FilesChanged:          querier.N(rsSnapshot.Summary.FilesChanged),
			FilesUnmodified:       querier.N(rsSnapshot.Summary.FilesUnmodified),
			TotalFilesProcessed:   querier.N(rsSnapshot.Summary.TotalFilesProcessed),
			TotalBytesProcessed:   querier.N(rsSnapshot.Summary.TotalBytesProcessed),
			DirsNew:               querier.N(rsSnapshot.Summary.DirsNew),
			DirsChanged:           querier.N(rsSnapshot.Summary.DirsChanged),
			DirsUnmodified:        querier.N(rsSnapshot.Summary.DirsUnmodified),
			TotalDirsProcessed:    querier.N(rsSnapshot.Summary.TotalDirsProcessed),
			TotalDirsizeProcessed: querier.N(rsSnapshot.Summary.TotalDirsizeProcessed),
			DataBlobs:             querier.N(rsSnapshot.Summary.DataBlobs),
			TreeBlobs:             querier.N(rsSnapshot.Summary.TreeBlobs),
			DataAdded:             querier.N(rsSnapshot.Summary.DataAdded),
			DataAddedPacked:       querier.N(rsSnapshot.Summary.DataAddedPacked),
			DataAddedFiles:        querier.N(rsSnapshot.Summary.DataAddedFiles),
			DataAddedFilesPacked:  querier.N(rsSnapshot.Summary.DataAddedFilesPacked),
			DataAddedTrees:        querier.N(rsSnapshot.Summary.DataAddedTrees),
			DataAddedTreesPacked:  querier.N(rsSnapshot.Summary.DataAddedTreesPacked),
			BackupStart:           querier.N(rsSnapshot.Summary.BackupStart),
			BackupEnd:             querier.N(rsSnapshot.Summary.BackupEnd),
			BackupDuration:        querier.N(float32(rsSnapshot.Summary.BackupDuration)),
			TotalDuration:         querier.N(float32(rsSnapshot.Summary.TotalDuration)),
		}
		err = snapshot.Rustic.Create(q)
		if err != nil {
			return nil, err
		}
	}
	return &snapshot, nil
}

func (r *Reconciler) ReconcileDeleteVolumeProvider(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, volumes map[int64]*dmodel.VolumeWithAttachment, snapshots map[int64]*dmodel.VolumeSnapshot) error {
	return nil
}
