package boxes

import (
	"context"
	"fmt"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
)

func (s *BoxesServer) restListComposeProjects(c context.Context, i *huma_utils.IdByPath) (*huma_utils.List[models.BoxComposeProject], error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	err := s.checkBoxToken(c, i.Id)
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

func (s *BoxesServer) validateBoxSpec(ctx context.Context, box *dmodel.Box) error {
	boxSpec, err := s.buildBoxSpec(ctx, box)
	if err != nil {
		return err
	}

	err = boxSpec.ValidateComposeProjects(ctx)
	if err != nil {
		return huma.Error400BadRequest(err.Error(), err)
	}

	return nil
}

type restCreateComposeProjectInput struct {
	huma_utils.IdByPath
	huma_utils.JsonBody[models.CreateBoxComposeProject]
}

func (s *BoxesServer) restCreateComposeProject(c context.Context, i *restCreateComposeProjectInput) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	err = s.createComposeProject(c, box, i.Body)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, box)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

func (s *BoxesServer) createComposeProject(c context.Context, box *dmodel.Box, req models.CreateBoxComposeProject) error {
	q := querier2.GetQuerier(c)

	err := util.CheckName(req.Name)
	if err != nil {
		return err
	}
	if strings.HasPrefix(req.Name, "dboxed-") {
		return huma.Error400BadRequest("'dboxed-' is a reserved internal prefix and can't be used")
	}

	bcps, err := dmodel.ListBoxComposeProjects(q, box.ID)
	if err != nil {
		return err
	}
	for _, bcp := range bcps {
		if bcp.Name == req.Name {
			return huma.Error400BadRequest(fmt.Sprintf("compose project with name %s already exists", req.Name))
		}
	}

	cp := dmodel.BoxComposeProject{
		BoxID:          box.ID,
		Name:           req.Name,
		ComposeProject: req.ComposeProject,
	}
	err = cp.Create(q)
	if err != nil {
		return err
	}

	err = s.validateBoxSpec(c, box)
	if err != nil {
		return err
	}

	return nil
}

type restUpdateComposeProjectInput struct {
	huma_utils.IdByPath
	ComposeName string `path:"composeName"`
	huma_utils.JsonBody[models.UpdateBoxComposeProject]
}

func (s *BoxesServer) restUpdateComposeProject(c context.Context, i *restUpdateComposeProjectInput) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	cp, err := dmodel.GetBoxComposeProjectByName(q, box.ID, i.ComposeName)
	if err != nil {
		return nil, err
	}

	err = cp.UpdateComposeProject(q, i.Body.ComposeProject)
	if err != nil {
		return nil, err
	}

	err = s.validateBoxSpec(c, box)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, box)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

type restDeleteComposeProjectInput struct {
	Id          int64  `path:"id"`
	ComposeName string `path:"composeName"`
}

func (s *BoxesServer) restDeleteComposeProject(c context.Context, i *restDeleteComposeProjectInput) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	cp, err := dmodel.GetBoxComposeProjectByName(q, box.ID, i.ComposeName)
	if err != nil {
		return nil, err
	}

	err = querier2.DeleteOneByFields[dmodel.BoxComposeProject](q, map[string]any{
		"box_id": box.ID,
		"name":   cp.Name,
	})
	if err != nil {
		return nil, err
	}

	err = s.validateBoxSpec(c, box)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, box)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}
