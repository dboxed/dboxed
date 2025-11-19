package auth_middleware

import (
	"context"

	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/models"
)

func IsAdminUser(ctx context.Context, u *models.User) bool {
	cfg := config.GetConfig(ctx)

	for _, au := range cfg.Auth.Oidc.AdminUsers {
		if au.ID != nil && *au.ID == u.ID {
			return true
		}
		if au.Username != nil && *au.Username == u.Username {
			return true
		}
	}
	return false
}
