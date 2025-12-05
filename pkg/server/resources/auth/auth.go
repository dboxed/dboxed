package auth

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/auth_middleware"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/huma_metadata"
)

const TokenPrefix = "dt_"

type AuthHandler struct {
	authInfo     models.AuthInfo
	oidcProvider *oidc.Provider
}

func NewAuthHandler(authInfo models.AuthInfo, oidcProvider *oidc.Provider) *AuthHandler {
	h := &AuthHandler{
		authInfo:     authInfo,
		oidcProvider: oidcProvider,
	}

	return h
}

func (s *AuthHandler) Init(api huma.API) error {
	huma.Get(api, "/v1/auth/current-user", s.restCurrentUser)
	huma.Get(api, "/v1/auth/current-token", s.restCurrentToken,
		huma_utils.MetadataModifier(huma_metadata.AllowAnyToken, true),
	)
	huma.Get(api, "/v1/auth/info", s.restInfo, huma_utils.MetadataModifier(huma_metadata.SkipAuth, true))

	return nil
}

func (s *AuthHandler) restInfo(ctx context.Context, input *struct{}) (*huma_utils.JsonBody[models.AuthInfo], error) {
	return huma_utils.NewJsonBody(s.authInfo), nil
}

func (s *AuthHandler) restCurrentUser(ctx context.Context, input *struct{}) (*huma_utils.JsonBody[models.User], error) {
	u := auth_middleware.GetUser(ctx)
	if u == nil {
		return nil, huma.Error404NotFound("no user")
	}
	return huma_utils.NewJsonBody(*u), nil
}

func (s *AuthHandler) restCurrentToken(ctx context.Context, input *struct{}) (*huma_utils.JsonBody[models.Token], error) {
	t := auth_middleware.GetToken(ctx)
	if t == nil {
		return nil, huma.Error404NotFound("no token")
	}
	t2 := *t
	t2.Token = nil
	return huma_utils.NewJsonBody(t2), nil
}
