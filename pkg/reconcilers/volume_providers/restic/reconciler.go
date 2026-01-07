package restic

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path"
	"strings"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/s3utils"
	"github.com/dboxed/dboxed/pkg/volume/restic"
	"github.com/minio/minio-go/v7"
)

type Reconciler struct {
}

func (r *Reconciler) buildResticEnv(vp *dmodel.VolumeProvider, b *dmodel.S3Bucket) []string {
	repo := fmt.Sprintf("%s/%s", b.Endpoint, b.Bucket)
	if vp.Restic.StoragePrefix.Valid && vp.Restic.StoragePrefix.V != "" {
		repo += "/" + vp.Restic.StoragePrefix.V
	}

	env := []string{
		fmt.Sprintf("RESTIC_REPOSITORY=s3:%s", repo),
		fmt.Sprintf("RESTIC_PASSWORD=%s", vp.Restic.Password.V),
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", b.AccessKeyId),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", b.SecretAccessKey),
	}
	if b.DeterminedRegion != nil {
		env = append(env, fmt.Sprintf("AWS_DEFAULT_REGION=%s", *b.DeterminedRegion))
	}
	return env
}

func (r *Reconciler) initResticRepo(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, b *dmodel.S3Bucket, c *minio.Client) base.ReconcileResult {
	_, err := c.StatObject(ctx, b.Bucket,
		path.Join(vp.Restic.StoragePrefix.V, "config"),
		minio.StatObjectOptions{})
	if err == nil {
		return base.ReconcileResult{}
	}
	var err2 minio.ErrorResponse
	if !errors.As(err, &err2) || err2.Code != minio.NoSuchKey {
		return base.ErrorWithMessage(err, "failed to determine if restic repo is already initialized: %s", err.Error())
	}

	log.InfoContext(ctx, "initializing restic repo")
	err = restic.RunInit(ctx, r.buildResticEnv(vp, b), restic.InitOpts{
		NoCache: true,
	})
	if err != nil {
		return base.ErrorWithMessage(err, "failed to initialize restic repo: %s", err.Error())
	}

	return base.ReconcileResult{}
}

func (r *Reconciler) listResticSnapshotIds(ctx context.Context, vp *dmodel.VolumeProvider, b *dmodel.S3Bucket, c *minio.Client) ([]string, base.ReconcileResult) {
	prefix := path.Join(vp.Restic.StoragePrefix.V, "snapshots") + "/"
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

func (r *Reconciler) deleteResticSnapshot(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, snapshotId string) base.ReconcileResult {
	log.InfoContext(ctx, "deleting restic snapshot", slog.Any("rsSnapshotId", snapshotId))

	b, c, err := s3utils.BuildS3ClientFromId(ctx, *vp.Restic.S3BucketID)
	if err != nil {
		return base.ErrorWithMessage(err, "failed building S3 client: %s", err.Error())
	}

	prefix := path.Join(vp.Restic.StoragePrefix.V, "snapshots") + "/"

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

func (r *Reconciler) ReconcileVolumeProvider(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, dbVolumes map[string]*dmodel.VolumeWithJoins, dbSnapshots map[string]*dmodel.VolumeSnapshot) base.ReconcileResult {
	dbSnapshotsByResticId := map[string]*dmodel.VolumeSnapshot{}
	for _, s := range dbSnapshots {
		dbSnapshotsByResticId[s.Restic.SnapshotId.V] = s
	}
	dbVolumesByUuuid := map[string]*dmodel.VolumeWithJoins{}
	for _, v := range dbVolumes {
		dbVolumesByUuuid[v.ID] = v
	}

	b, c, err := s3utils.BuildS3ClientFromId(ctx, *vp.Restic.S3BucketID)
	if err != nil {
		return base.ErrorWithMessage(err, "failed building S3 client: %s", err.Error())
	}

	result := r.initResticRepo(ctx, log, vp, b, c)
	if result.ExitReconcile() {
		return result
	}

	rsSnapshotIds, result := r.listResticSnapshotIds(ctx, vp, b, c)
	if result.ExitReconcile() {
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
			slog.Any("snapshotResticId", s.Restic.SnapshotId.V),
		)

		_, ok := rsSnapshotIdSet[s.Restic.SnapshotId.V]
		if !ok {
			log.InfoContext(ctx, "snapshot vanished from restic, marking for deletion in DB")
			err := querier.Transaction(ctx, func(ctx context.Context) (bool, error) {
				q := querier.GetQuerier(ctx)
				v, ok := dbVolumes[s.VolumedID.V]
				if ok {
					if v.LatestSnapshotId != nil && *v.LatestSnapshotId == s.ID {
						log.InfoContext(ctx, "snapshot was the latest snapshot, resetting latest snapshot field")
						err := v.UpdateLatestSnapshot(q, nil)
						if err != nil {
							return false, err
						}
					}
				}
				err := dmodel.SoftDeleteByStruct(q, s)
				if err != nil {
					return false, err
				}
				return true, nil
			})
			if err != nil {
				return base.InternalError(err)
			}
		}
	}

	return base.ReconcileResult{}
}

func (r *Reconciler) ReconcileDeleteSnapshot(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, dbVolume *dmodel.Volume, dbSnapshot *dmodel.VolumeSnapshot) base.ReconcileResult {
	result := r.deleteResticSnapshot(ctx, log, vp, dbSnapshot.Restic.SnapshotId.V)
	if result.ExitReconcile() {
		return result
	}
	return base.ReconcileResult{}
}

func (r *Reconciler) ReconcileDeleteVolume(ctx context.Context, log *slog.Logger, vp *dmodel.VolumeProvider, dbVolume *dmodel.VolumeWithJoins) base.ReconcileResult {
	// we assume that all snapshots have been deleted already
	return base.ReconcileResult{}
}
