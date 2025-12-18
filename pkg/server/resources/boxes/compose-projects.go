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

func (s *BoxesServer) restListComposeProjects(c context.Context, i *huma_utils.IdByPath) (*huma_utils.List[models.BoxComposeProject], error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	err := auth_middleware.CheckTokenAccess(c, dmodel.TokenTypeBox, i.Id)
	if err != nil {
		return nil, err
	}

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	cps, err := dmodel.ListBoxComposeProjects(q, box.ID)
	if err != nil {
		return nil, err
	}

	var ret []models.BoxComposeProject
	for _, a := range cps {
		ma := models.BoxComposeProjectFromDB(a)
		ret = append(ret, *ma)
	}

	return huma_utils.NewList(ret, len(ret)), nil
}

type restCreateComposeProjectInput struct {
	huma_utils.IdByPath
	huma_utils.JsonBody[models.CreateBoxComposeProject]
}

func (s *BoxesServer) restCreateComposeProject(c context.Context, i *restCreateComposeProjectInput) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}
	if err = s.checkNormalBoxMod(box); err != nil {
		return nil, err
	}

	err = boxes_utils.CreateComposeProject(c, box, i.Body)
	if err != nil {
		return nil, err
	}

	err = dmodel.BumpChangeSeq(q, box)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

type restUpdateComposeProjectInput struct {
	huma_utils.IdByPath
	ComposeName string `path:"composeName"`
	huma_utils.JsonBody[models.UpdateBoxComposeProject]
}

func (s *BoxesServer) restUpdateComposeProject(c context.Context, i *restUpdateComposeProjectInput) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	if err = s.checkNormalBoxMod(box); err != nil {
		return nil, err
	}

	err = boxes_utils.UpdateComposeProject(c, box, i.ComposeName, i.Body.ComposeProject)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

type restDeleteComposeProjectInput struct {
	Id          string `path:"id"`
	ComposeName string `path:"composeName"`
}

func (s *BoxesServer) restDeleteComposeProject(c context.Context, i *restDeleteComposeProjectInput) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}
	if err = s.checkNormalBoxMod(box); err != nil {
		return nil, err
	}

	err = boxes_utils.DeleteComposeProject(c, box, i.ComposeName)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}
