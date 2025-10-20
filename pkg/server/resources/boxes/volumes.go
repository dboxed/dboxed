package boxes

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
)

func (s *BoxesServer) restListAttachedVolumes(c context.Context, i *huma_utils.IdByPath) (*huma_utils.List[models.VolumeAttachment], error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	err := s.checkBoxToken(c, i.Id)
	if err != nil {
		return nil, err
	}

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	attachments, err := dmodel.ListBoxVolumeAttachments(q, box.ID)
	if err != nil {
		return nil, err
	}

	var ret []models.VolumeAttachment
	for _, a := range attachments {
		ma := models.VolumeAttachmentFromDB(a.BoxVolumeAttachment, &a.Volume, nil)
		ret = append(ret, ma)
	}

	return huma_utils.NewList(ret, len(ret)), nil
}

func (s *BoxesServer) checkVolumeAttachmentParams(rootUid *int64, rootGid *int64, rootMode *string) error {
	if rootMode != nil {
		mode, err := strconv.ParseInt(*rootMode, 8, 32)
		if err != nil {
			return huma.Error400BadRequest("invalid root_mode", err)
		}
		if mode&^int64(boxspec.AllowedModeMask) != 0 {
			return huma.Error400BadRequest("invalid root_mode", err)
		}
	}
	if rootUid != nil && *rootUid < 0 {
		return huma.Error400BadRequest("invalid root_uid", nil)
	}
	if rootGid != nil && *rootGid < 0 {
		return huma.Error400BadRequest("invalid root_gid", nil)
	}
	return nil
}

type restAttachVolumeInput struct {
	huma_utils.IdByPath
	huma_utils.JsonBody[models.AttachVolumeRequest]
}

func (s *BoxesServer) restAttachVolume(c context.Context, i *restAttachVolumeInput) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	err = s.attachVolume(c, box, i.Body)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, box)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

func (s *BoxesServer) attachVolume(c context.Context, box *dmodel.Box, req models.AttachVolumeRequest) error {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	volume, err := dmodel.GetVolumeById(q, &w.ID, req.VolumeId, true)
	if err != nil {
		return err
	}

	err = s.checkVolumeAttachmentParams(req.RootUid, req.RootGid, req.RootMode)
	if err != nil {
		return err
	}

	slog.InfoContext(c, "attaching volume to box", slog.Any("boxId", box.ID), slog.Any("volumeId", req.VolumeId))

	attachment := &dmodel.BoxVolumeAttachment{
		BoxId:    querier2.N(box.ID),
		VolumeId: querier2.N(volume.ID),
		RootUid:  querier2.N(int64(0)),
		RootGid:  querier2.N(int64(0)),
		RootMode: querier2.N("0777"),
	}

	if req.RootUid != nil {
		attachment.RootUid = querier2.N(*req.RootUid)
	}
	if req.RootGid != nil {
		attachment.RootGid = querier2.N(*req.RootGid)
	}
	if req.RootMode != nil {
		attachment.RootMode = querier2.N(*req.RootMode)
	}

	err = attachment.Create(q)
	if err != nil {
		return err
	}

	err = s.validateBoxSpec(c, box, false)
	if err != nil {
		return err
	}

	return nil
}

type restUpdateAttachedVolumeInput struct {
	Id       int64 `path:"id"`
	VolumeId int64 `path:"volumeId"`
	huma_utils.JsonBody[models.UpdateVolumeAttachmentRequest]
}

func (s *BoxesServer) restUpdateAttachedVolume(c context.Context, i *restUpdateAttachedVolumeInput) (*huma_utils.JsonBody[models.VolumeAttachment], error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	volume, err := dmodel.GetVolumeById(q, &w.ID, i.VolumeId, true)
	if err != nil {
		return nil, err
	}

	attachment, err := dmodel.GetBoxVolumeAttachment(q, box.ID, volume.ID)
	if err != nil {
		return nil, err
	}

	err = s.checkVolumeAttachmentParams(i.Body.RootUid, i.Body.RootGid, i.Body.RootMode)
	if err != nil {
		return nil, err
	}

	err = attachment.Update(q, i.Body.RootUid, i.Body.RootGid, i.Body.RootMode)
	if err != nil {
		return nil, err
	}

	err = s.validateBoxSpec(c, box, false)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, box)
	if err != nil {
		return nil, err
	}

	ret := models.VolumeAttachmentFromDB(attachment.BoxVolumeAttachment, &attachment.Volume, nil)
	return huma_utils.NewJsonBody(ret), nil
}

type restDetachVolumeInput struct {
	Id       int64 `path:"id"`
	VolumeId int64 `path:"volumeId"`
}

func (s *BoxesServer) restDetachVolume(c context.Context, i *restDetachVolumeInput) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	volume, err := dmodel.GetVolumeById(q, &w.ID, i.VolumeId, true)
	if err != nil {
		return nil, err
	}

	err = querier2.DeleteOneByFields[dmodel.BoxVolumeAttachment](q, map[string]any{
		"box_id":    box.ID,
		"volume_id": volume.ID,
	})
	if err != nil {
		return nil, err
	}

	err = s.validateBoxSpec(c, box, false)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, box)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}
