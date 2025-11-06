package token

import (
	"context"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/google/uuid"
)

type TokenCommands struct {
	Create CreateCmd `cmd:"" help:"Create a token"`
	Get    GetCmd    `cmd:"" help:"Get a token"`
	List   ListCmd   `cmd:"" help:"List tokens" aliases:"ls"`
	Delete DeleteCmd `cmd:"" help:"Delete a token" aliases:"rm,delete"`
}

func getToken(ctx context.Context, c *baseclient.Client, token string) (*models.Token, error) {
	c2 := clients.TokenClient{Client: c}
	if uuid.Validate(token) == nil {
		v, err := c2.GetTokenById(ctx, token)
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
