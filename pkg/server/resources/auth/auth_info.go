package auth

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/models"
)

func BuildAuthProvider(ctx context.Context, c config.Config) (*models.AuthInfo, *oidc.Provider, error) {
	var ret models.AuthInfo
	var oidcProvider *oidc.Provider

	if c.Auth.Oidc != nil {
		ret.OidcIssuerUrl = c.Auth.Oidc.IssuerUrl
		ret.OidcClientId = c.Auth.Oidc.ClientId

		var err error
		oidcProvider, err = oidc.NewProvider(ctx, c.Auth.Oidc.IssuerUrl)
		if err != nil {
			return nil, nil, err
		}
	}
	if ret.OidcIssuerUrl == "" {
		return nil, nil, fmt.Errorf("missing oidc issuer url")
	}
	if oidcProvider == nil {
		return nil, nil, fmt.Errorf("failed to create oidc provider")
	}

	return &ret, oidcProvider, nil
}
