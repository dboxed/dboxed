package boxes_utils

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/server/box_spec_utils"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
)

func BuildBoxSpec(c context.Context, box *dmodel.Box, withNetwork bool) (*boxspec.BoxSpec, error) {
	q := querier.GetQuerier(c)

	var network *dmodel.Network
	if withNetwork {
		if box.NetworkID != nil {
			var err error
			network, err = dmodel.GetNetworkById(q, &box.OwnedByWorkspace.WorkspaceID, *box.NetworkID, true)
			if err != nil {
				return nil, err
			}
		}
	}

	file, err := box_spec_utils.BuildBoxSpec(c, box, network)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func ValidateBoxSpec(ctx context.Context, box *dmodel.Box, withNetwork bool) error {
	boxSpec, err := BuildBoxSpec(ctx, box, withNetwork)
	if err != nil {
		return err
	}

	err = boxSpec.ValidateComposeProjects(ctx)
	if err != nil {
		return huma.Error400BadRequest(err.Error(), err)
	}

	return nil
}
