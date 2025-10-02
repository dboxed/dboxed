package volumes

import (
	"context"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
)

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

	err = dmodel.SoftDeleteByIds[*dmodel.VolumeSnapshot](q, &w.ID, i.SnapshotId)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}
