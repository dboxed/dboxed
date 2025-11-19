package workspaces

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/auth_middleware"
	config2 "github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/huma_metadata"
	"github.com/dboxed/dboxed/pkg/util"
)

type WorkspacesServer struct {
	api huma.API

	Middleware *auth_middleware.WorkspaceMiddleware
}

func New() *WorkspacesServer {
	s := &WorkspacesServer{}
	s.Middleware = &auth_middleware.WorkspaceMiddleware{
		GetWorkspace: s.getWorkspaceById,
	}
	return s
}

func (s *WorkspacesServer) Init(api huma.API) error {
	s.api = api

	skipWorkspaceModifier := huma_utils.MetadataModifier(huma_metadata.SkipWorkspace, true)
	allowTokensModifier := huma_utils.MetadataModifier(huma_metadata.AllowTokensWithWorkspace, true)

	huma.Post(s.api, "/v1/workspaces", s.restCreateWorkspace, skipWorkspaceModifier)
	huma.Get(s.api, "/v1/workspaces", s.restListWorkspaces, skipWorkspaceModifier, allowTokensModifier)
	huma.Get(s.api, "/v1/workspaces/{workspaceId}", s.restGetWorkspace, skipWorkspaceModifier, allowTokensModifier)
	huma.Delete(s.api, "/v1/workspaces/{workspaceId}", s.restDeleteWorkspace, skipWorkspaceModifier)

	huma.Get(s.api, "/v1/admin/workspaces", s.restAdminListWorkspaces, skipWorkspaceModifier, huma_metadata.NeedAdminModifier())

	return nil
}

func (s *WorkspacesServer) restCreateWorkspace(ctx context.Context, i *huma_utils.JsonBody[models.CreateWorkspace]) (*huma_utils.JsonBody[models.Workspace], error) {
	q := querier2.GetQuerier(ctx)
	config := config2.GetConfig(ctx)

	user := auth_middleware.MustGetUser(ctx)

	err := util.CheckName(i.Body.Name)
	if err != nil {
		return nil, err
	}

	w := &dmodel.Workspace{
		Name: i.Body.Name,
		Access: []dmodel.WorkspaceAccess{
			{UserId: user.ID},
		},
	}
	err = w.Create(q)
	if err != nil {
		return nil, err
	}

	wq := dmodel.WorkspaceQuotas{
		WorkspaceId: w.ID,
		MaxLogBytes: config.DefaultWorkspaceQuotas.MaxLogBytes.Bytes,
	}
	err = wq.Create(q)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, w)
	if err != nil {
		return nil, err
	}

	return huma_utils.NewJsonBody(models.WorkspaceFromDB(*w)), nil
}

func (s *WorkspacesServer) restListWorkspaces(ctx context.Context, i *struct{}) (*huma_utils.List[models.Workspace], error) {
	return s.doRestListWorkspaces(ctx, false)
}

func (s *WorkspacesServer) restAdminListWorkspaces(ctx context.Context, i *struct{}) (*huma_utils.List[models.Workspace], error) {
	return s.doRestListWorkspaces(ctx, true)
}

func (s *WorkspacesServer) doRestListWorkspaces(ctx context.Context, asAdmin bool) (*huma_utils.List[models.Workspace], error) {
	q := querier2.GetQuerier(ctx)
	user := auth_middleware.GetUser(ctx)
	token := auth_middleware.GetToken(ctx)

	var workspaces []dmodel.Workspace
	if user != nil {
		var err error
		if asAdmin {
			workspaces, err = dmodel.ListWorkspaces(q, nil, true)
		} else {
			workspaces, err = dmodel.ListWorkspaces(q, &user.ID, true)
		}
		if err != nil {
			return nil, err
		}
	} else if token != nil {
		// return only the single workspace assigned to the token
		w, err := dmodel.GetWorkspaceById(q, token.Workspace, true)
		if err != nil {
			return nil, err
		}
		workspaces = append(workspaces, *w)
	} else {
		return nil, huma.Error401Unauthorized("missing user/token")
	}

	var ret []models.Workspace
	for _, w := range workspaces {
		ret = append(ret, models.WorkspaceFromDB(w))
	}
	return huma_utils.NewList(ret, len(ret)), nil
}

func (s *WorkspacesServer) restGetWorkspace(ctx context.Context, i *models.WorkspaceIdByPath) (*huma_utils.JsonBody[models.Workspace], error) {
	w, err := s.Middleware.CheckWorkspaceAccess(ctx, i.WorkspaceId, false)
	if err != nil {
		return nil, err
	}
	return huma_utils.NewJsonBody(*w), nil
}

func (s *WorkspacesServer) restDeleteWorkspace(ctx context.Context, i *models.WorkspaceIdByPath) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(ctx)

	w, err := s.Middleware.CheckWorkspaceAccess(ctx, i.WorkspaceId, true)
	if err != nil {
		return nil, err
	}

	err = dmodel.SoftDeleteWithConstraints[*dmodel.Workspace](q, map[string]any{
		"id": w.ID,
	}, nil)
	if err != nil {
		return nil, err
	}
	err = dmodel.AddChangeTrackingForId[*dmodel.Workspace](q, w.ID)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

func (s *WorkspacesServer) getWorkspaceById(ctx context.Context, id string) (*models.Workspace, error) {
	q := querier2.GetQuerier(ctx)
	w, err := dmodel.GetWorkspaceById(q, id, true)
	if err != nil {
		if querier2.IsSqlNotFoundError(err) {
			return nil, huma.Error404NotFound("workspace not found")
		}
		return nil, err
	}
	m := models.WorkspaceFromDB(*w)
	return &m, nil
}
