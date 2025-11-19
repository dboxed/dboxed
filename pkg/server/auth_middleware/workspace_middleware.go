package auth_middleware

import (
	"context"
	"errors"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/huma_metadata"
)

type GetWorkspaceFunc func(ctx context.Context, id string) (*models.Workspace, error)

type WorkspaceMiddleware struct {
	GetWorkspace GetWorkspaceFunc
}

func (s *WorkspaceMiddleware) CheckWorkspaceAccess(ctx context.Context, id string, onlyWorkspaceToken bool) (*models.Workspace, error) {
	user := GetUser(ctx)
	token := GetToken(ctx)

	w, err := s.GetWorkspace(ctx, id)
	if err != nil {
		return nil, err
	}

	if user != nil {
		if !user.IsAdmin {
			if !slices.ContainsFunc(w.Access, func(access models.WorkspaceAccess) bool {
				return access.User.ID == user.ID
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

func (s *WorkspaceMiddleware) WorkspaceMiddleware(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		s.workspaceMiddleware(api, ctx, next)
	}
}

func (s *WorkspaceMiddleware) workspaceMiddleware(api huma.API, ctx huma.Context, next func(huma.Context)) {
	if huma_utils.HasMetadataTrue(ctx, huma_metadata.SkipWorkspace) ||
		huma_utils.HasMetadataTrue(ctx, huma_metadata.SkipAuth) {
		next(ctx)
		return
	}

	workspaceId := ctx.Param("workspaceId")
	if workspaceId == "" {
		huma.WriteErr(api, ctx, http.StatusBadRequest, "missing workspace id")
		return
	}

	w, err := s.CheckWorkspaceAccess(ctx.Context(), workspaceId, false)
	if err != nil {
		var err2 huma.StatusError
		if errors.As(err, &err2) {
			huma.WriteErr(api, ctx, err2.GetStatus(), err.Error(), err)
		} else {
			huma.WriteErr(api, ctx, http.StatusForbidden, err.Error(), err)
		}
		return
	}

	ctx = huma.WithValue(ctx, "workspace", w)

	next(ctx)
}

func GetWorkspace(ctx context.Context) *models.Workspace {
	return global.MustGet[*models.Workspace](ctx, "workspace")
}
