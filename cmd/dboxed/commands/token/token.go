package token

import (
	"context"
	"strconv"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type TokenCommands struct {
	Create TokenCreateCmd `cmd:"" help:"Create a token"`
	Get    TokenGetCmd    `cmd:"" help:"Get a token"`
	List   TokenListCmd   `cmd:"" help:"List tokens"`
	Delete TokenDeleteCmd `cmd:"" help:"Delete a token"`
}

func getToken(ctx context.Context, c *baseclient.Client, token string) (*models.Token, error) {
	c2 := clients.TokenClient{Client: c}
	id, err := strconv.ParseInt(token, 10, 64)
	if err == nil {
		v, err := c2.GetTokenById(ctx, id)
		if err != nil {
			return nil, err
		}
		return v, nil
	} else {
		v, err := c2.GetTokenByName(ctx, token)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}
