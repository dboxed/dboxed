package boxes

import (
	"context"

	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/server/auth_middleware"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/resources/boxes_utils"
)

func (s *BoxesServer) restGetBoxSpec(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[boxspec.BoxSpec], error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	err := s.checkBoxToken(c, i.Id)
	if err != nil {
		return nil, err
	}

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	ret, err := boxes_utils.BuildBoxSpec(c, box, true)
	if err != nil {
		return nil, err
	}
	return huma_utils.NewJsonBody(*ret), nil
}
