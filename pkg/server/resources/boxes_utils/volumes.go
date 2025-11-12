package boxes_utils

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/models"
)

func AttachVolume(c context.Context, box *dmodel.Box, req models.AttachVolumeRequest) error {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	volume, err := dmodel.GetVolumeById(q, &w.ID, req.VolumeId, true)
	if err != nil {
		return err
	}

	err = CheckVolumeAttachmentParams(req.RootUid, req.RootGid, req.RootMode)
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

	err = ValidateBoxSpec(c, box, false)
	if err != nil {
		return err
	}

	return nil
}

func CheckVolumeAttachmentParams(rootUid *int64, rootGid *int64, rootMode *string) error {
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
