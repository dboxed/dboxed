package boxes

import (
	"context"

	"github.com/dboxed/dboxed/pkg/server/auth_middleware"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/boxes_utils"
)

func (s *BoxesServer) restListAttachedVolumes(c context.Context, i *huma_utils.IdByPath) (*huma_utils.List[models.VolumeAttachment], error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

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
		var mountStatus *dmodel.VolumeMountStatus
		if a.Volume.MountId != nil {
			mountStatus, err = dmodel.GetVolumeMountStatusById(q, a.VolumeId.V, *a.Volume.MountId)
			if err != nil && !querier2.IsSqlNotFoundError(err) {
				return nil, err
			}
		}
		ma := models.VolumeAttachmentFromDB(a.BoxVolumeAttachment, &a.Volume, nil, mountStatus)
		ret = append(ret, ma)
	}

	return huma_utils.NewList(ret, len(ret)), nil
}

type restAttachVolumeInput struct {
	huma_utils.IdByPath
	huma_utils.JsonBody[models.AttachVolumeRequest]
}

func (s *BoxesServer) restAttachVolume(c context.Context, i *restAttachVolumeInput) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}
	if err = s.checkNormalBoxMod(box); err != nil {
		return nil, err
	}

	err = boxes_utils.AttachVolume(c, box, i.Body)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, box)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

type restUpdateAttachedVolumeInput struct {
	Id       string `path:"id"`
	VolumeId string `path:"volumeId"`
	huma_utils.JsonBody[models.UpdateVolumeAttachmentRequest]
}

func (s *BoxesServer) restUpdateAttachedVolume(c context.Context, i *restUpdateAttachedVolumeInput) (*huma_utils.JsonBody[models.VolumeAttachment], error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}
	if err = s.checkNormalBoxMod(box); err != nil {
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

	err = boxes_utils.CheckVolumeAttachmentParams(i.Body.RootUid, i.Body.RootGid, i.Body.RootMode)
	if err != nil {
		return nil, err
	}

	err = attachment.Update(q, i.Body.RootUid, i.Body.RootGid, i.Body.RootMode)
	if err != nil {
		return nil, err
	}

	err = boxes_utils.ValidateBoxSpec(c, box, false)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, box)
	if err != nil {
		return nil, err
	}

	ret := models.VolumeAttachmentFromDB(attachment.BoxVolumeAttachment, &attachment.Volume, nil, volume.MountStatus)
	return huma_utils.NewJsonBody(ret), nil
}

type restDetachVolumeInput struct {
	Id       string `path:"id"`
	VolumeId string `path:"volumeId"`
}

func (s *BoxesServer) restDetachVolume(c context.Context, i *restDetachVolumeInput) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}
	if err = s.checkNormalBoxMod(box); err != nil {
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

	err = boxes_utils.ValidateBoxSpec(c, box, false)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, box)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}
