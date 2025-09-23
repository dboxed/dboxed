package auth

import (
	"fmt"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

func GetAuthorizationToken(ctx huma.Context) (string, error) {
	tokenString := ctx.Header("Authorization")
	if tokenString == "" {
		return "", fmt.Errorf("missing authentication token")
	}

	// The token should be prefixed with "Bearer "
	tokenParts := strings.Split(tokenString, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authentication token")
	}

	return tokenParts[1], nil
}
