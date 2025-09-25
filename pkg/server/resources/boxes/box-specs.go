package boxes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/server/box_spec_utils"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/resources/auth"
)

func (s *BoxesServer) restGetBoxSpec(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[boxspec.BoxFile], error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)
	token := auth.GetToken(c)

	if token != nil && token.BoxID != nil {
		if *token.BoxID != i.Id {
			return nil, huma.Error403Forbidden("no access to box")
		}
	}

	box, err := dmodel.GetBoxById(q, &w.ID, *token.BoxID, true)
	if err != nil {
		return nil, err
	}

	ret, err := s.buildBoxSpec(c, box)
	if err != nil {
		return nil, err
	}
	return huma_utils.NewJsonBody(*ret), nil
}

func (s *BoxesServer) buildBoxSpec(c context.Context, box *dmodel.Box) (*boxspec.BoxFile, error) {
	q := querier.GetQuerier(c)

	var network *dmodel.Network
	if box.NetworkID != nil {
		var err error
		network, err = dmodel.GetNetworkById(q, &box.OwnedByWorkspace.WorkspaceID, *box.NetworkID, true)
		if err != nil {
			return nil, err
		}
	}

	file, err := box_spec_utils.BuildBoxSpec(c, box, network)
	if err != nil {
		return nil, err
	}

	return file, nil
}
