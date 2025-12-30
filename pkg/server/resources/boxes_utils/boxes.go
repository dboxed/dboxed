package boxes_utils

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
)

func CreateBox(c context.Context, workspaceId string, body models.CreateBox, boxType dmodel.BoxType) (*dmodel.Box, error) {
	q := querier2.GetQuerier(c)

	err := util.CheckName(body.Name)
	if err != nil {
		return nil, err
	}

	box := &dmodel.Box{
		OwnedByWorkspace: dmodel.OwnedByWorkspace{
			WorkspaceID: workspaceId,
		},
		Name:    body.Name,
		BoxType: boxType,

		Enabled: true,
	}

	var network *dmodel.Network
	if body.Network != nil {
		network, err = dmodel.GetNetworkById(q, &workspaceId, *body.Network, true)
		if err != nil {
			return nil, err
		}
		box.NetworkID = &network.ID
		box.NetworkType = &network.Type
	}

	err = box.Create(q)
	if err != nil {
		return nil, err
	}

	if network != nil {
		switch network.Type {
		case dmodel.NetworkTypeNetbird:
			box.Netbird = &dmodel.BoxNetbird{
				ID: querier2.N(box.ID),
			}
			err = box.Netbird.Create(q)
			if err != nil {
				return nil, err
			}
		default:
			return nil, huma.Error400BadRequest("unknown network type")
		}
	}

	for _, va := range body.VolumeAttachments {
		err = AttachVolume(c, box, va)
		if err != nil {
			return nil, err
		}
	}

	for _, cp := range body.ComposeProjects {
		err = CreateComposeProject(c, box, cp)
		if err != nil {
			return nil, err
		}
	}

	if network != nil {
		err = dmodel.BumpChangeSeq(q, network)
		if err != nil {
			return nil, err
		}
	}

	return box, nil
}

func DeleteBox(c context.Context, workspaceId string, boxId string) error {
	q := querier2.GetQuerier(c)
	box, err := dmodel.GetBoxById(q, &workspaceId, boxId, true)
	if err != nil {
		return err
	}
	err = dmodel.SoftDeleteWithConstraintsByIds[*dmodel.Box](q, &workspaceId, boxId)
	if err != nil {
		return err
	}
	if box.NetworkID != nil {
		err = dmodel.BumpChangeSeqForId[*dmodel.Network](q, *box.NetworkID)
		if err != nil {
			return err
		}
	}
	return nil
}
