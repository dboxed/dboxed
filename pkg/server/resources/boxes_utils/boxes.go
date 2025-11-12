package boxes_utils

import (
	"context"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
)

func CreateBox(c context.Context, body models.CreateBox, boxType global.BoxType) (*dmodel.Box, string, error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	err := util.CheckName(body.Name)
	if err != nil {
		return nil, err.Error(), nil
	}

	var networkId *string
	var networkType *string
	if body.Network != nil {
		var network *dmodel.Network
		network, err = dmodel.GetNetworkById(q, &w.ID, *body.Network, true)
		if err != nil {
			return nil, "", err
		}
		networkId = &network.ID
		networkType = &network.Type
	}

	box := &dmodel.Box{
		OwnedByWorkspace: dmodel.OwnedByWorkspace{
			WorkspaceID: w.ID,
		},
		Name:    body.Name,
		BoxType: string(boxType),

		DboxedVersion: "nightly",
		DesiredState:  "up",

		NetworkID:   networkId,
		NetworkType: networkType,
	}

	err = box.Create(q)
	if err != nil {
		return nil, "", err
	}

	sandboxStatus := dmodel.BoxSandboxStatus{
		ID: querier2.N(box.ID),
	}
	err = sandboxStatus.Create(q)
	if err != nil {
		return nil, "", err
	}

	if networkId != nil {
		switch global.NetworkType(*networkType) {
		case global.NetworkNetbird:
			box.Netbird = &dmodel.BoxNetbird{
				ID: querier2.N(box.ID),
			}
			err = box.Netbird.Create(q)
			if err != nil {
				return nil, "", err
			}
		default:
			return nil, "unknown network type", nil
		}
	}

	for _, va := range body.VolumeAttachments {
		err = AttachVolume(c, box, va)
		if err != nil {
			return nil, "", err
		}
	}

	for _, cp := range body.ComposeProjects {
		err = CreateComposeProject(c, box, cp)
		if err != nil {
			return nil, "", err
		}
	}

	err = dmodel.AddChangeTracking(q, box)
	if err != nil {
		return nil, "", err
	}

	return box, "", nil
}
