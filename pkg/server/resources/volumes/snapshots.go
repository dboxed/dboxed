package volumes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type restCreateSnapshotInput struct {
	huma_utils.IdByPath
	huma_utils.JsonBody[models.CreateVolumeSnapshot]
}

func (s *VolumeServer) restCreateSnapshot(ctx context.Context, i *restCreateSnapshotInput) (*huma_utils.JsonBody[models.VolumeSnapshot], error) {
	q := querier.GetQuerier(ctx)
	w := global.GetWorkspace(ctx)

	v, err := dmodel.GetVolumeById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	snapshot := dmodel.VolumeSnapshot{
		OwnedByWorkspace: dmodel.OwnedByWorkspace{
			WorkspaceID: w.ID,
		},
		VolumeProviderID: v.VolumeProviderID,
		VolumedID:        i.Id,
		LockID:           i.Body.LockID,
	}

	err = snapshot.Create(q)
	if err != nil {
		return nil, err
	}

	if v.Rustic != nil {
		if i.Body.Rustic == nil {
			return nil, huma.Error400BadRequest("missing rustic in body")
		}
		snapshot.Rustic = &dmodel.VolumeSnapshotRustic{
			ID:                    querier.N(snapshot.ID),
			SnapshotId:            querier.N(i.Body.Rustic.SnapshotId),
			SnapshotTime:          querier.N(i.Body.Rustic.SnapshotTime),
			ParentSnapshotId:      i.Body.Rustic.ParentSnapshotId,
			Hostname:              querier.N(i.Body.Rustic.Hostname),
			FilesNew:              querier.N(i.Body.Rustic.FilesNew),
			FilesChanged:          querier.N(i.Body.Rustic.FilesChanged),
			FilesUnmodified:       querier.N(i.Body.Rustic.FilesUnmodified),
			TotalFilesProcessed:   querier.N(i.Body.Rustic.TotalFilesProcessed),
			TotalBytesProcessed:   querier.N(i.Body.Rustic.TotalBytesProcessed),
			DirsNew:               querier.N(i.Body.Rustic.DirsNew),
			DirsChanged:           querier.N(i.Body.Rustic.DirsChanged),
			DirsUnmodified:        querier.N(i.Body.Rustic.DirsUnmodified),
			TotalDirsProcessed:    querier.N(i.Body.Rustic.TotalDirsProcessed),
			TotalDirsizeProcessed: querier.N(i.Body.Rustic.TotalDirsizeProcessed),
			DataBlobs:             querier.N(i.Body.Rustic.DataBlobs),
			TreeBlobs:             querier.N(i.Body.Rustic.TreeBlobs),
			DataAdded:             querier.N(i.Body.Rustic.DataAdded),
			DataAddedPacked:       querier.N(i.Body.Rustic.DataAddedPacked),
			DataAddedFiles:        querier.N(i.Body.Rustic.DataAddedFiles),
			DataAddedFilesPacked:  querier.N(i.Body.Rustic.DataAddedFilesPacked),
			DataAddedTrees:        querier.N(i.Body.Rustic.DataAddedTrees),
			DataAddedTreesPacked:  querier.N(i.Body.Rustic.DataAddedTreesPacked),
			BackupStart:           querier.N(i.Body.Rustic.BackupStart),
			BackupEnd:             querier.N(i.Body.Rustic.BackupEnd),
			BackupDuration:        querier.N(i.Body.Rustic.BackupDuration),
			TotalDuration:         querier.N(i.Body.Rustic.TotalDuration),
		}
		err = snapshot.Rustic.Create(q)
		if err != nil {
			return nil, err
		}
	}

	if i.Body.IsLatest {
		err = v.UpdateLatestSnapshot(q, &snapshot.ID)
		if err != nil {
			return nil, err
		}
	}

	return huma_utils.NewJsonBody(models.VolumeSnapshotFromDB(snapshot)), nil
}

func (s *VolumeServer) restListSnapshots(ctx context.Context, i *huma_utils.IdByPath) (*huma_utils.List[models.VolumeSnapshot], error) {
	q := querier.GetQuerier(ctx)
	w := global.GetWorkspace(ctx)

	l, err := dmodel.ListVolumeSnapshotsForVolume(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	var ret []models.VolumeSnapshot
	for _, r := range l {
		mm := models.VolumeSnapshotFromDB(r)
		ret = append(ret, mm)
	}
	return huma_utils.NewList(ret, len(ret)), nil
}

type snapshotIdByPath struct {
	huma_utils.IdByPath
	SnapshotId int64 `path:"snapshotId"`
}

func (s *VolumeServer) restGetSnapshot(ctx context.Context, i *snapshotIdByPath) (*huma_utils.JsonBody[models.VolumeSnapshot], error) {
	q := querier.GetQuerier(ctx)
	w := global.GetWorkspace(ctx)

	v, err := dmodel.GetVolumeSnapshotById(q, &w.ID, &i.Id, i.SnapshotId, true)
	if err != nil {
		return nil, err
	}

	m := models.VolumeSnapshotFromDB(*v)
	return huma_utils.NewJsonBody(m), nil
}

func (s *VolumeServer) restDeleteSnapshot(ctx context.Context, i *snapshotIdByPath) (*huma_utils.Empty, error) {
	q := querier.GetQuerier(ctx)
	w := global.GetWorkspace(ctx)

	_, err := dmodel.GetVolumeSnapshotById(q, &w.ID, &i.Id, i.SnapshotId, true)
	if err != nil {
		return nil, err
	}

	err = dmodel.SoftDeleteWithConstraintsByIds[dmodel.VolumeSnapshot](q, &w.ID, i.SnapshotId)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}
