package boxes

import (
	"context"

	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/server/box_spec_utils"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
)

func (s *BoxesServer) restGetBoxSpec(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[boxspec.BoxSpec], error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	err := s.checkBoxToken(c, i.Id)
	if err != nil {
		return nil, err
	}

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	ret, err := s.buildBoxSpec(c, box, true)
	if err != nil {
		return nil, err
	}
	return huma_utils.NewJsonBody(*ret), nil
}

func (s *BoxesServer) buildBoxSpec(c context.Context, box *dmodel.Box, withNetwork bool) (*boxspec.BoxSpec, error) {
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
