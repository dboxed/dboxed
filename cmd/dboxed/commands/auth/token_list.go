package auth

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type TokenListCmd struct{}

func (cmd *TokenListCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.TokenClient{Client: c}

	tokens, err := c2.ListTokens(ctx)
	if err != nil {
		return err
	}

	for _, token := range tokens {
		slog.Info("token", slog.Any("id", token.ID), slog.Any("name", token.Name), slog.Any("created_at", token.CreatedAt))
	}

	return nil
}
