package dboxed_specs

import (
	"context"
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/auth_middleware"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/kluctl/kluctl/lib/git/types"
)

type DboedSpecsServer struct {
}

func New() *DboedSpecsServer {
	return &DboedSpecsServer{}
}

func (s *DboedSpecsServer) Init(rootGroup huma.API, workspacesGroup huma.API) error {
	huma.Post(workspacesGroup, "/dboxed-specs", s.restCreateDboxedSpec)
	huma.Get(workspacesGroup, "/dboxed-specs", s.restListDboxedSpecs)
	huma.Get(workspacesGroup, "/dboxed-specs/{id}", s.restGetDboxedSpec)
	huma.Patch(workspacesGroup, "/dboxed-specs/{id}", s.restUpdateDboxedSpec)
	huma.Delete(workspacesGroup, "/dboxed-specs/{id}", s.restDeleteDboxedSpec)

	return nil
}

func (s *DboedSpecsServer) restCreateDboxedSpec(c context.Context, i *huma_utils.JsonBody[models.CreateDboxedSpec]) (*huma_utils.JsonBody[models.DboxedSpec], error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	_, err := types.ParseGitUrl(i.Body.GitUrl)
	if err != nil {
		return nil, huma.Error400BadRequest(fmt.Sprintf("invalid git url: %s", err.Error()), err)
	}

	gs := &dmodel.DboxedSpec{
		OwnedByWorkspace: dmodel.OwnedByWorkspace{
			WorkspaceID: w.ID,
		},
		GitUrl:   i.Body.GitUrl,
		Subdir:   i.Body.Subdir,
		SpecFile: i.Body.SpecFile,
	}
	gs.SetGitRef(i.Body.GitRef)

	err = gs.Create(q)
	if err != nil {
		return nil, err
	}

	m := models.DboxedSpecFromDB(*gs)
	return huma_utils.NewJsonBody(m), nil
}

func (s *DboedSpecsServer) restListDboxedSpecs(c context.Context, i *struct{}) (*huma_utils.List[models.DboxedSpec], error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	l, err := dmodel.ListDboxedSpecsForWorkspace(q, w.ID, true)
	if err != nil {
		return nil, err
	}

	var ret []models.DboxedSpec
	for _, gs := range l {
		ret = append(ret, models.DboxedSpecFromDB(gs))
	}
	return huma_utils.NewList(ret, len(ret)), nil
}

func (s *DboedSpecsServer) restGetDboxedSpec(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[models.DboxedSpec], error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	gs, err := dmodel.GetDboxedSpecById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	m := models.DboxedSpecFromDB(*gs)
	return huma_utils.NewJsonBody(m), nil
}

type restUpdateDboxedSpecInput struct {
	huma_utils.IdByPath
	huma_utils.JsonBody[models.UpdateDboxedSpec]
}

func (s *DboedSpecsServer) restUpdateDboxedSpec(c context.Context, i *restUpdateDboxedSpecInput) (*huma_utils.JsonBody[models.DboxedSpec], error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	if i.Body.GitUrl != nil {
		_, err := types.ParseGitUrl(*i.Body.GitUrl)
		if err != nil {
			return nil, huma.Error400BadRequest(fmt.Sprintf("invalid git url: %s", err.Error()), err)
		}
	}

	gs, err := dmodel.GetDboxedSpecById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	err = gs.Update(q, i.Body.GitUrl, &i.Body.GitRef, i.Body.Subdir, i.Body.SpecFile)
	if err != nil {
		return nil, err
	}

	err = dmodel.BumpChangeSeq(q, gs)
	if err != nil {
		return nil, err
	}

	m := models.DboxedSpecFromDB(*gs)
	return huma_utils.NewJsonBody(m), nil
}

func (s *DboedSpecsServer) restDeleteDboxedSpec(c context.Context, i *huma_utils.IdByPath) (*huma_utils.Empty, error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	err := dmodel.SoftDeleteWithConstraintsByIds[*dmodel.DboxedSpec](q, &w.ID, i.Id)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}
