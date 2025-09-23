package boxes

import (
	"context"

	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/server/box_spec_utils"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type restGetBoxSpecInput struct {
	Token string `query:"token"`
}

func (s *BoxesServer) restGetBoxSpec(c context.Context, i *restGetBoxSpecInput) (*huma_utils.JsonBody[boxspec.BoxFile], error) {
	q := querier.GetQuerier(c)

	token, sig, err := models.TokenFromStr(i.Token)
	if err != nil {
		return nil, err
	}

	box, err := dmodel.GetBoxById(q, &token.WorkspaceId, token.BoxId, true)
	if err != nil {
		return nil, err
	}

	err = token.Verify(box.Nkey, sig)
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
