package auth

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
)

type TokenGetCmd struct {
	Id string `help:"Token ID" required:"" arg:""`
}

func (cmd *TokenGetCmd) Run() error {
	ctx := context.Background()

	c, err := baseclient.FromClientAuthFile()
	if err != nil {
		return err
	}

	c2 := &clients.TokenClient{Client: c}

	token, err := getToken(ctx, c2, cmd.Id)
	if err != nil {
		return err
	}

	slog.Info("token", slog.Any("id", token.ID), slog.Any("name", token.Name), slog.Any("created_at", token.CreatedAt))

	return nil
}
