package auth_middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
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

type AuthMiddleware struct {
	authConfig   config.AuthConfig
	authInfo     models.AuthInfo
	oidcProvider *oidc.Provider
	enableDbUser bool
}

func NewAuthMiddleware(authConfig config.AuthConfig, authInfo models.AuthInfo, oidcProvider *oidc.Provider, enableDbUser bool) *AuthMiddleware {
	h := &AuthMiddleware{
		authConfig:   authConfig,
		authInfo:     authInfo,
		oidcProvider: oidcProvider,
		enableDbUser: enableDbUser,
	}

	return h
}

// verifyIDToken verifies that an *oauth2.Token is a valid *oidc.IDToken.
func (s *AuthMiddleware) verifyIDToken(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
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

func (s *AuthMiddleware) buildUserFromIDToken(ctx context.Context, idToken *oidc.IDToken) (*models.User, error) {
	var claims jwt.MapClaims
	err := idToken.Claims(&claims)
	if err != nil {
		return nil, err
	}

	sub, err := claims.GetSubject()
	if err != nil {
		return nil, err
	}

	usernameClaim := s.authConfig.Oidc.UsernameClaim
	emailClaim := s.authConfig.Oidc.EMailClaim
	fullNameClaim := s.authConfig.Oidc.FullNameClaim

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

	u := &models.User{
		ID:       sub,
		Username: *username,
		EMail:    email,
		FullName: name,
		Avatar:   avatar,
	}
	u.IsAdmin = IsAdminUser(s.authConfig, u)

	return u, nil
}

func (s *AuthMiddleware) AuthMiddleware(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		s.authMiddleware(api, ctx, next)
	}
}

func (s *AuthMiddleware) authMiddleware(api huma.API, ctx huma.Context, next func(huma.Context)) {
	if huma_utils.HasMetadataTrue(ctx, huma_metadata.SkipAuth) {
		next(ctx)
		return
	}

	authz, err := GetAuthorizationTokenFromHuma(ctx)
	if err != nil {
		_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, err.Error(), err)
		return
	}

	ctx = huma.WithValue(ctx, "authorizationToken", authz)

	if strings.HasPrefix(authz, TokenPrefix) {
		if huma_utils.HasMetadataTrue(ctx, huma_metadata.NeedAdmin) {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "token not allowed")
			return
		}
		token, err := s.checkDboxedToken(ctx, authz)
		if err != nil {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, err.Error(), err)
			return
		}
		ctx = huma.WithValue(ctx, "token", token)
	} else {
		user, err := s.checkIdToken(ctx, authz)
		if err != nil {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, err.Error(), err)
			return
		}
		ctx = huma.WithValue(ctx, "user", user)

		err = s.updateDBUser(ctx, user)
		if err != nil {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, err.Error(), err)
			return
		}

		if huma_utils.HasMetadataTrue(ctx, huma_metadata.NeedAdmin) {
			if !user.IsAdmin {
				_ = huma.WriteErr(api, ctx, http.StatusForbidden, "must be admin")
				return
			}
		}
	}

	next(ctx)
}

func (s *AuthMiddleware) checkDboxedToken(ctx huma.Context, authz string) (*models.Token, error) {
	q := querier.GetQuerier(ctx.Context())
	t, err := dmodel.GetTokenByToken(q, authz)
	if err != nil {
		return nil, err
	}

	ok := false
	if huma_utils.HasMetadataTrue(ctx, huma_metadata.AllowAnyToken) {
		ok = true
	} else {
		switch t.Type {
		case dmodel.TokenTypeWorkspace:
			ok = huma_utils.HasMetadataTrue(ctx, huma_metadata.AllowWorkspaceToken)
		case dmodel.TokenTypeMachine:
			ok = huma_utils.HasMetadataTrue(ctx, huma_metadata.AllowMachineToken)
		case dmodel.TokenTypeBox:
			ok = huma_utils.HasMetadataTrue(ctx, huma_metadata.AllowBoxToken)
		case dmodel.TokenTypeLoadBalancer:
			ok = huma_utils.HasMetadataTrue(ctx, huma_metadata.AllowLoadBalancerToken)
		}
	}
	if !ok {
		return nil, fmt.Errorf("%s token not allowed", t.Type)
	}

	m := models.TokenFromDB(*t, false)
	return &m, nil
}

func (s *AuthMiddleware) checkIdToken(ctx huma.Context, authz string) (*models.User, error) {
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

func (s *AuthMiddleware) updateDBUser(ctx huma.Context, user *models.User) error {
	if !s.enableDbUser {
		return nil
	}

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
		um := models.UserFromDB(*dbUser)
		um.IsAdmin = IsAdminUser(s.authConfig, &um)
		needUpdate = !reflect.DeepEqual(um, *user)
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

func GetAuthorizationToken(ctx context.Context) *string {
	tokenI := ctx.Value("authorizationToken")
	if tokenI == nil {
		return nil
	}
	token, ok := tokenI.(string)
	if !ok {
		return nil
	}
	return &token
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

func CheckTokenAccess(ctx context.Context, expectedTokenType dmodel.TokenType, resourceId string) error {
	token := GetToken(ctx)
	if token == nil {
		user := GetUser(ctx)
		if user == nil {
			return huma.Error403Forbidden("missing user and token")
		}
		return nil
	}

	if token.Type == dmodel.TokenTypeWorkspace {
		return nil
	}
	if token.Type != expectedTokenType {
		return huma.Error403Forbidden(fmt.Sprintf("token does not have access to %s", expectedTokenType))
	}

	switch token.Type {
	case dmodel.TokenTypeMachine:
		if *token.MachineID != resourceId {
			return huma.Error403Forbidden("token does not have access to machine")
		}
		return nil
	case dmodel.TokenTypeBox:
		if *token.BoxID != resourceId {
			return huma.Error403Forbidden("token does not have access to box")
		}
		return nil
	case dmodel.TokenTypeLoadBalancer:
		if *token.LoadBalancerId != resourceId {
			return huma.Error403Forbidden("token does not have access to load-balancer")
		}
		return nil
	default:
		return huma.Error403Forbidden("token does not have access to resource")
	}
}
