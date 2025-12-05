package auth_middleware

import (
	"context"
	"errors"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/huma_metadata"
)

type WorkspaceMiddleware struct {
}

func CheckWorkspaceAccess(ctx context.Context, w *models.Workspace, onlyWorkspaceToken bool) error {
	user := GetUser(ctx)
	token := GetToken(ctx)

	if user != nil {
		if !user.IsAdmin {
			if !slices.ContainsFunc(w.Access, func(access models.WorkspaceAccess) bool {
				return access.User.ID == user.ID
			}) {
				return huma.Error403Forbidden("access to workspace not allowed")
			}
		}
	} else if token != nil {
		if onlyWorkspaceToken && token.Type != dmodel.TokenTypeWorkspace {
			return huma.Error403Forbidden("access to workspace not allowed")
		}
		if token.Workspace != w.ID {
			return huma.Error403Forbidden("access to workspace not allowed")
		}
	} else {
		return huma.Error403Forbidden("access to workspace not allowed")
	}

	return nil
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

	q := querier2.GetQuerier(ctx.Context())
	w, err := dmodel.GetWorkspaceById(q, workspaceId, true)
	if err != nil {
		if querier2.IsSqlNotFoundError(err) {
			huma.WriteErr(api, ctx, http.StatusNotFound, "workspace not found")
		} else {
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	wm := models.WorkspaceFromDB(*w)

	err = CheckWorkspaceAccess(ctx.Context(), &wm, false)
	if err != nil {
		var err2 huma.StatusError
		if errors.As(err, &err2) {
			huma.WriteErr(api, ctx, err2.GetStatus(), err.Error(), err)
		} else {
			huma.WriteErr(api, ctx, http.StatusForbidden, err.Error(), err)
		}
		return
	}

	ctx = huma.WithValue(ctx, "workspace", &wm)

	next(ctx)
}

func GetWorkspace(ctx context.Context) *models.Workspace {
	return global.MustGet[*models.Workspace](ctx, "workspace")
}
