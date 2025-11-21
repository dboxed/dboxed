package auth_middleware

import (
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/models"
)

func IsAdminUser(authConfig config.AuthConfig, u *models.User) bool {
	for _, au := range authConfig.Oidc.AdminUsers {
		if au.ID != nil && *au.ID == u.ID {
			return true
		}
		if au.Username != nil && *au.Username == u.Username {
			return true
		}
	}
	return false
}
