package rustic

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/dboxed/dboxed/pkg/server/db/dbutils"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/volume/rustic"
	"github.com/dboxed/dboxed/pkg/volume/volume_backup"
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

func (r *Reconciler) listRusticSnapshots(ctx context.Context, vp *dmodel.VolumeProvider) ([]rustic.Snapshot, error) {
	config, err := r.buildRusticConfig(vp)
	if err != nil {
		return nil, err
	}

	groups, err := rustic.RunSnapshots(ctx, *config, nil)
	if err != nil {
		return nil, err
	}

	expectedTags := volume_backup.BuildBackupTags(vp.ID, nil, nil, nil)

	var ret []rustic.Snapshot
	for _, g := range groups {
		for _, s := range g.Snapshots {
			allFound := true
			for _, tag := range expectedTags {
				if !slices.Contains(s.Tags, tag) {
					allFound = false
					break
				}
			}
			if allFound {
				ret = append(ret, s)
			}
		}
	}

	return ret, nil
}

func (r *Reconciler) deleteRusticSnapshots(ctx context.Context, vp *dmodel.VolumeProvider, snapshotIds []string) error {
	slog.InfoContext(ctx, "deleting rustic snapshots", slog.Any("rsSnapshotIds", snapshotIds))

	config, err := r.buildRusticConfig(vp)
	if err != nil {
		return err
	}

	err = rustic.RunForget(ctx, *config, rustic.ForgetOpts{
		SnapshotIds: snapshotIds,
	})
	if err != nil {
		return err
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

	rsSnapshots, err := r.listRusticSnapshots(ctx, vp)
	if err != nil {
		return fmt.Errorf("failed to list rustic snapshots: %w", err)
	}
	rsSnapshotsByRusticId := map[string]*rustic.Snapshot{}
	for _, s := range rsSnapshots {
		rsSnapshotsByRusticId[s.Id] = &s
	}

	var rsSnapshotsToDelete []string
	for _, s := range rsSnapshotsByRusticId {
		volumeUuid := r.getTagValue(s.Tags, "dboxed-volume-uuid")
		v := dbVolumesByUuuid[volumeUuid]

		if v == nil || v.DeletedAt.Valid {
			rsSnapshotsToDelete = append(rsSnapshotsToDelete, s.Id)
			continue
		}

		dbSnapshot, ok := dbSnapshotsByRusticId[s.Id]
		if !ok {
			err = r.createDBSnapshot(ctx, log, vp, &v.Volume, s)
			if err != nil {
				return fmt.Errorf("failed to create snapshot: %w", err)
			}
			continue
		}
		if dbSnapshot.DeletedAt.Valid {
			rsSnapshotsToDelete = append(rsSnapshotsToDelete, s.Id)
		}
	}

	if len(rsSnapshotsToDelete) != 0 {
		err = r.deleteRusticSnapshots(ctx, vp, rsSnapshotsToDelete)
		if err != nil {
			return fmt.Errorf("failed to delete snapshots: %w", err)
		}
		for _, s := range rsSnapshotsToDelete {
			delete(rsSnapshotsByRusticId, s)
		}
	}

	for _, s := range dbSnapshots {
		if s.DeletedAt.Valid {
			continue
		}
		log := log.With(
			slog.Any("volumeId", s.VolumedID),
			slog.Any("snapshotId", s.ID),
			slog.Any("snapshotRusticId", s.Rustic.SnapshotId.V),
		)

		_, ok := rsSnapshotsByRusticId[s.Rustic.SnapshotId.V]
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

func (r *Reconciler) createDBSnapshot(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, v *dmodel.Volume, rsSnapshot *rustic.Snapshot) error {
	return dbutils.RunInTx(ctx, func(ctx context.Context) error {
		return r.createDBSnapshotInTx(ctx, log, vp, v, rsSnapshot)
	})
}

func (r *Reconciler) createDBSnapshotInTx(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, v *dmodel.Volume, rsSnapshot *rustic.Snapshot) error {
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
		return err
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
			return err
		}
	}
	return nil
}

func (r *Reconciler) ReconcileDeleteVolumeProvider(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, volumes map[int64]*dmodel.VolumeWithAttachment, snapshots map[int64]*dmodel.VolumeSnapshot) error {
	return nil
}
