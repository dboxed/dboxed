package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"slices"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/huma_metadata"
	"github.com/golang-jwt/jwt/v5"
)

const TokenPrefix = "dt_"

type AuthHandler struct {
	api huma.API

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

func (s *AuthHandler) Init(ctx context.Context, api huma.API) error {
	s.api = api

	huma.Get(s.api, "/v1/auth/current-user", s.restCurrentUser)
	huma.Get(s.api, "/v1/auth/current-token", s.restCurrentToken,
		huma_utils.MetadataModifier(huma_metadata.AllowWorkspaceToken, true),
		huma_utils.MetadataModifier(huma_metadata.AllowBoxToken, true),
	)
	huma.Get(s.api, "/v1/auth/info", s.restInfo, huma_utils.MetadataModifier(huma_metadata.SkipAuth, true))

	return nil
}

func (s *AuthHandler) restInfo(ctx context.Context, input *struct{}) (*huma_utils.JsonBody[models.AuthInfo], error) {
	return huma_utils.NewJsonBody(s.authInfo), nil
}

func (s *AuthHandler) restCurrentUser(ctx context.Context, input *struct{}) (*huma_utils.JsonBody[models.User], error) {
	u := GetUser(ctx)
	if u == nil {
		return nil, huma.Error404NotFound("no user")
	}
	return huma_utils.NewJsonBody(*u), nil
}

func (s *AuthHandler) restCurrentToken(ctx context.Context, input *struct{}) (*huma_utils.JsonBody[models.Token], error) {
	t := GetToken(ctx)
	if t == nil {
		return nil, huma.Error404NotFound("no token")
	}
	t2 := *t
	t2.Token = nil
	return huma_utils.NewJsonBody(t2), nil
}

// verifyIDToken verifies that an *oauth2.Token is a valid *oidc.IDToken.
func (s *AuthHandler) verifyIDToken(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	oidcConfig := &oidc.Config{
		ClientID: s.authInfo.OidcClientId,
	}

	return s.oidcProvider.Verifier(oidcConfig).Verify(ctx, rawIDToken)
}

func getClaimValue[T any](m jwt.MapClaims, n string, missingOk bool) (*T, error) {
	i, ok := m[n]
	if !ok {
		if missingOk {
			return nil, nil
		}
		return nil, fmt.Errorf("missing %s claim", n)
	}
	v, ok := i.(T)
	if !ok {
		return nil, fmt.Errorf("invalid %s claim", n)
	}
	return &v, nil
}

func (s *AuthHandler) buildUserFromIDToken(ctx context.Context, idToken *oidc.IDToken) (*models.User, error) {
	cfg := config.GetConfig(ctx)

	var claims jwt.MapClaims
	err := idToken.Claims(&claims)
	if err != nil {
		return nil, err
	}

	sub, err := claims.GetSubject()
	if err != nil {
		return nil, err
	}

	usernameClaim := cfg.Auth.Oidc.UsernameClaim
	emailClaim := cfg.Auth.Oidc.EMailClaim
	fullNameClaim := cfg.Auth.Oidc.FullNameClaim

	if usernameClaim == "" {
		usernameClaim = "email"
	}
	if emailClaim == "" {
		emailClaim = "email"
	}
	if fullNameClaim == "" {
		fullNameClaim = "name"
	}

	username, err := getClaimValue[string](claims, usernameClaim, false)
	if err != nil {
		return nil, err
	}
	email, err := getClaimValue[string](claims, emailClaim, true)
	if err != nil {
		return nil, err
	}
	name, err := getClaimValue[string](claims, fullNameClaim, true)
	if err != nil {
		return nil, err
	}

	avatar, err := getClaimValue[string](claims, "avatar", true)
	if err != nil {
		return nil, err
	}

	isAdmin := false
	if slices.Contains(cfg.Auth.AdminUsers, sub) {
		isAdmin = true
	}

	return &models.User{
		ID:       sub,
		Username: *username,
		EMail:    email,
		FullName: name,
		Avatar:   avatar,
		IsAdmin:  isAdmin,
	}, nil
}

func (s *AuthHandler) AuthMiddleware(ctx huma.Context, next func(huma.Context)) {
	if huma_utils.HasMetadataTrue(ctx, huma_metadata.SkipAuth) {
		next(ctx)
		return
	}

	authz, err := GetAuthorizationToken(ctx)
	if err != nil {
		_ = huma.WriteErr(s.api, ctx, http.StatusUnauthorized, err.Error(), err)
		return
	}

	if strings.HasPrefix(authz, TokenPrefix) {
		if huma_utils.HasMetadataTrue(ctx, huma_metadata.NeedAdmin) {
			_ = huma.WriteErr(s.api, ctx, http.StatusUnauthorized, "token not allowed")
			return
		}
		token, err := s.checkDboxedToken(ctx, authz)
		if err != nil {
			_ = huma.WriteErr(s.api, ctx, http.StatusUnauthorized, err.Error(), err)
			return
		}
		ctx = huma.WithValue(ctx, "token", token)
	} else {
		user, err := s.checkIdToken(ctx, authz)
		if err != nil {
			_ = huma.WriteErr(s.api, ctx, http.StatusUnauthorized, err.Error(), err)
			return
		}
		ctx = huma.WithValue(ctx, "user", user)

		err = s.updateDBUser(ctx, user)
		if err != nil {
			_ = huma.WriteErr(s.api, ctx, http.StatusUnauthorized, err.Error(), err)
			return
		}

		if huma_utils.HasMetadataTrue(ctx, huma_metadata.NeedAdmin) {
			if !user.IsAdmin {
				_ = huma.WriteErr(s.api, ctx, http.StatusForbidden, "must be admin")
				return
			}
		}
	}

	next(ctx)
}

func (s *AuthHandler) checkDboxedToken(ctx huma.Context, authz string) (*models.Token, error) {
	q := querier.GetQuerier(ctx.Context())
	t, err := dmodel.GetTokenByToken(q, authz)
	if err != nil {
		return nil, err
	}

	allowTokensWithWorkspace := huma_utils.HasMetadataTrue(ctx, huma_metadata.AllowTokensWithWorkspace)

	if t.ForWorkspace && !allowTokensWithWorkspace && !huma_utils.HasMetadataTrue(ctx, huma_metadata.AllowWorkspaceToken) {
		return nil, fmt.Errorf("workspace token not allowed")
	}
	if t.BoxID != nil && !allowTokensWithWorkspace && !huma_utils.HasMetadataTrue(ctx, huma_metadata.AllowBoxToken) {
		return nil, fmt.Errorf("box token not allowed")
	}

	m := models.TokenFromDB(*t, false)
	return &m, nil
}

func (s *AuthHandler) checkIdToken(ctx huma.Context, authz string) (*models.User, error) {
	idToken, err := s.verifyIDToken(ctx.Context(), authz)
	if err != nil {
		return nil, err
	}
	user, err := s.buildUserFromIDToken(ctx.Context(), idToken)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthHandler) updateDBUser(ctx huma.Context, user *models.User) error {
	q := querier2.GetQuerier(ctx.Context())

	newDbUser := dmodel.User{
		ID:       user.ID,
		Username: user.Username,
		EMail:    user.EMail,
		FullName: user.FullName,
		Avatar:   user.Avatar,
	}

	needUpdate := false
	dbUser, err := dmodel.GetUserById(q, user.ID)
	if err != nil {
		if !querier2.IsSqlNotFoundError(err) {
			return err
		}
		needUpdate = true
	} else {
		needUpdate = !reflect.DeepEqual(models.UserFromDB(*dbUser, user.IsAdmin), *user)
	}
	if !needUpdate {
		return nil
	}
	slog.InfoContext(ctx.Context(), "updating user in DB", slog.Any("user", *user))

	err = newDbUser.CreateOrUpdate(q)
	if err != nil {
		return err
	}
	return nil
}

func GetUser(ctx context.Context) *models.User {
	userI := ctx.Value("user")
	if userI == nil {
		return nil
	}
	user, ok := userI.(*models.User)
	if !ok {
		return nil
	}
	return user
}

func MustGetUser(ctx context.Context) models.User {
	user := GetUser(ctx)
	if user == nil {
		panic("missing user")
	}
	return *user
}

func GetToken(ctx context.Context) *models.Token {
	tokenI := ctx.Value("token")
	if tokenI == nil {
		return nil
	}
	token, ok := tokenI.(*models.Token)
	if !ok {
		return nil
	}
	return token
}
