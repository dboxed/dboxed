package auth_middleware

import (
	"fmt"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

func GetAuthorizationTokenFromHuma(ctx huma.Context) (string, error) {
	tokenString := ctx.Header("Authorization")
	if tokenString == "" {
		tokenString = ctx.Query("token")
		if tokenString != "" {
			return tokenString, nil
		}

		return "", fmt.Errorf("missing authentication token")
	}

	// The token should be prefixed with "Bearer "
	tokenParts := strings.Split(tokenString, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authentication token")
	}

	return tokenParts[1], nil
}
