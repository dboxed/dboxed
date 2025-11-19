package workspaces

import (
	"context"
	"errors"
	"net/http"
	"slices"

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

type Workspaces struct {
	api huma.API
}

func New() *Workspaces {
	return &Workspaces{}
}

func (s *Workspaces) Init(api huma.API) error {
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

func (s *Workspaces) restCreateWorkspace(ctx context.Context, i *huma_utils.JsonBody[models.CreateWorkspace]) (*huma_utils.JsonBody[models.Workspace], error) {
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

func (s *Workspaces) restListWorkspaces(ctx context.Context, i *struct{}) (*huma_utils.List[models.Workspace], error) {
	return s.doRestListWorkspaces(ctx, false)
}

func (s *Workspaces) restAdminListWorkspaces(ctx context.Context, i *struct{}) (*huma_utils.List[models.Workspace], error) {
	return s.doRestListWorkspaces(ctx, true)
}

func (s *Workspaces) doRestListWorkspaces(ctx context.Context, asAdmin bool) (*huma_utils.List[models.Workspace], error) {
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

func (s *Workspaces) restGetWorkspace(ctx context.Context, i *models.WorkspaceIdByPath) (*huma_utils.JsonBody[models.Workspace], error) {
	w, err := s.checkWorkspaceAccess(ctx, i.WorkspaceId, false)
	if err != nil {
		return nil, err
	}
	return huma_utils.NewJsonBody(models.WorkspaceFromDB(*w)), nil
}

func (s *Workspaces) restDeleteWorkspace(ctx context.Context, i *models.WorkspaceIdByPath) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(ctx)

	w, err := s.checkWorkspaceAccess(ctx, i.WorkspaceId, true)
	if err != nil {
		return nil, err
	}

	err = dmodel.SoftDeleteWithConstraints[*dmodel.Workspace](q, map[string]any{
		"id": w.ID,
	}, nil)
	if err != nil {
		return nil, err
	}
	err = dmodel.AddChangeTracking(q, w)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

func (s *Workspaces) checkWorkspaceAccess(ctx context.Context, id string, onlyWorkspaceToken bool) (*dmodel.Workspace, error) {
	q := querier2.GetQuerier(ctx)
	user := auth_middleware.GetUser(ctx)
	token := auth_middleware.GetToken(ctx)

	w, err := dmodel.GetWorkspaceById(q, id, true)
	if err != nil {
		if querier2.IsSqlNotFoundError(err) {
			return nil, huma.Error404NotFound("workspace not found")
		}
		return nil, err
	}

	if user != nil {
		if !user.IsAdmin {
			if !slices.ContainsFunc(w.Access, func(access dmodel.WorkspaceAccess) bool {
				return access.UserId == user.ID
			}) {
				return nil, huma.Error403Forbidden("access to workspace not allowed")
			}
		}
	} else if token != nil {
		if onlyWorkspaceToken && !token.ForWorkspace {
			return nil, huma.Error403Forbidden("access to workspace not allowed")
		}
		if token.Workspace != id {
			return nil, huma.Error403Forbidden("access to workspace not allowed")
		}
	} else {
		return nil, huma.Error403Forbidden("access to workspace not allowed")
	}

	return w, nil
}

func (s *Workspaces) WorkspaceMiddleware(ctx huma.Context, next func(huma.Context)) {
	if huma_utils.HasMetadataTrue(ctx, huma_metadata.SkipWorkspace) ||
		huma_utils.HasMetadataTrue(ctx, huma_metadata.SkipAuth) {
		next(ctx)
		return
	}

	workspaceId := ctx.Param("workspaceId")
	if workspaceId == "" {
		huma.WriteErr(s.api, ctx, http.StatusBadRequest, "missing workspace id")
		return
	}

	w, err := s.checkWorkspaceAccess(ctx.Context(), workspaceId, false)
	if err != nil {
		var err2 huma.StatusError
		if errors.As(err, &err2) {
			huma.WriteErr(s.api, ctx, err2.GetStatus(), err.Error(), err)
		} else {
			huma.WriteErr(s.api, ctx, http.StatusForbidden, err.Error(), err)
		}
		return
	}

	ctx = huma.WithValue(ctx, "workspace", w)

	next(ctx)
}
