package auth

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/baseclient"
)

type TokenGetCmd struct {
	Token string `help:"Specify the token" required:"" arg:""`
}

func (cmd *TokenGetCmd) Run() error {
	ctx := context.Background()

	c, err := baseclient.FromClientAuthFile()
	if err != nil {
		return err
	}

	token, err := getToken(ctx, c, cmd.Token)
	if err != nil {
		return err
	}

	slog.Info("token", slog.Any("id", token.ID), slog.Any("name", token.Name), slog.Any("created_at", token.CreatedAt))

	return nil
}
