package token

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
)

type GetCmd struct {
	Token string `help:"Specify the token" required:"" arg:""`
}

func (cmd *GetCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
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
