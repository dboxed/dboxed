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

	var networkId *string
	var networkType *dmodel.NetworkType
	if body.Network != nil {
		var network *dmodel.Network
		network, err = dmodel.GetNetworkById(q, &workspaceId, *body.Network, true)
		if err != nil {
			return nil, err
		}
		networkId = &network.ID
		networkType = &network.Type
	}

	box := &dmodel.Box{
		OwnedByWorkspace: dmodel.OwnedByWorkspace{
			WorkspaceID: workspaceId,
		},
		Name:    body.Name,
		BoxType: boxType,

		Enabled: true,

		NetworkID:   networkId,
		NetworkType: networkType,
	}

	err = box.Create(q)
	if err != nil {
		return nil, err
	}

	sandboxStatus := dmodel.BoxSandboxStatus{
		ID: querier2.N(box.ID),
	}
	err = sandboxStatus.Create(q)
	if err != nil {
		return nil, err
	}

	if networkId != nil {
		switch *networkType {
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

	return box, nil
}

func DeleteBox(c context.Context, workspaceId string, boxId string) error {
	q := querier2.GetQuerier(c)
	err := dmodel.SoftDeleteWithConstraintsByIds[*dmodel.Box](q, &workspaceId, boxId)
	if err != nil {
		return err
	}
	return nil
}
