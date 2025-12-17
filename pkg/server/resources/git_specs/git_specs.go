package git_specs

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

type GitSpecsServer struct {
}

func New() *GitSpecsServer {
	return &GitSpecsServer{}
}

func (s *GitSpecsServer) Init(rootGroup huma.API, workspacesGroup huma.API) error {
	huma.Post(workspacesGroup, "/git-specs", s.restCreateGitSpec)
	huma.Get(workspacesGroup, "/git-specs", s.restListGitSpecs)
	huma.Get(workspacesGroup, "/git-specs/{id}", s.restGetGitSpec)
	huma.Patch(workspacesGroup, "/git-specs/{id}", s.restUpdateGitSpec)
	huma.Delete(workspacesGroup, "/git-specs/{id}", s.restDeleteGitSpec)

	return nil
}

func (s *GitSpecsServer) restCreateGitSpec(c context.Context, i *huma_utils.JsonBody[models.CreateGitSpec]) (*huma_utils.JsonBody[models.GitSpec], error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	_, err := types.ParseGitUrl(i.Body.GitUrl)
	if err != nil {
		return nil, huma.Error400BadRequest(fmt.Sprintf("invalid git url: %s", err.Error()), err)
	}

	gs := &dmodel.GitSpec{
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

	m := models.GitSpecFromDB(*gs)
	return huma_utils.NewJsonBody(m), nil
}

func (s *GitSpecsServer) restListGitSpecs(c context.Context, i *struct{}) (*huma_utils.List[models.GitSpec], error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	l, err := dmodel.ListGitSpecsForWorkspace(q, w.ID, true)
	if err != nil {
		return nil, err
	}

	var ret []models.GitSpec
	for _, gs := range l {
		ret = append(ret, models.GitSpecFromDB(gs))
	}
	return huma_utils.NewList(ret, len(ret)), nil
}

func (s *GitSpecsServer) restGetGitSpec(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[models.GitSpec], error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	gs, err := dmodel.GetGitSpecsById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	m := models.GitSpecFromDB(*gs)
	return huma_utils.NewJsonBody(m), nil
}

type restUpdateGitSpecInput struct {
	huma_utils.IdByPath
	huma_utils.JsonBody[models.UpdateGitSpec]
}

func (s *GitSpecsServer) restUpdateGitSpec(c context.Context, i *restUpdateGitSpecInput) (*huma_utils.JsonBody[models.GitSpec], error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	if i.Body.GitUrl != nil {
		_, err := types.ParseGitUrl(*i.Body.GitUrl)
		if err != nil {
			return nil, huma.Error400BadRequest(fmt.Sprintf("invalid git url: %s", err.Error()), err)
		}
	}

	gs, err := dmodel.GetGitSpecsById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	err = gs.Update(q, i.Body.GitUrl, &i.Body.GitRef, i.Body.Subdir, i.Body.SpecFile)
	if err != nil {
		return nil, err
	}

	m := models.GitSpecFromDB(*gs)
	return huma_utils.NewJsonBody(m), nil
}

func (s *GitSpecsServer) restDeleteGitSpec(c context.Context, i *huma_utils.IdByPath) (*huma_utils.Empty, error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	err := dmodel.SoftDeleteWithConstraintsByIds[*dmodel.GitSpec](q, &w.ID, i.Id)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}
