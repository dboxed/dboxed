package auth

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type TokenCreateCmd struct {
	Name string `help:"Specify the token name. Must be unique." required:""`
}

func (cmd *TokenCreateCmd) Run() error {
	ctx := context.Background()

	c, err := baseclient.FromClientAuthFile()
	if err != nil {
		return err
	}

	c2 := &clients.TokenClient{Client: c}

	req := models.CreateToken{
		Name: cmd.Name,
	}

	token, err := c2.CreateToken(ctx, req)
	if err != nil {
		return err
	}

	slog.Info("token created", slog.Any("id", token.ID), slog.Any("name", token.Name), slog.Any("token", token.TokenStr))

	return nil
}
