package auth

import (
	"context"

	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type TokenCmd struct {
	Create TokenCreateCmd `cmd:"" help:"Create a token"`
	Get    TokenGetCmd    `cmd:"" help:"Get a token"`
	List   TokenListCmd   `cmd:"" help:"List tokens"`
	Delete TokenDeleteCmd `cmd:"" help:"Delete a token"`
}

func getToken(ctx context.Context, c *clients.TokenClient, tokenId string) (*models.Token, error) {
	return c.GetTokenById(ctx, tokenId)
}
