package token

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type TokenDeleteCmd struct {
	Token string `help:"Specify the token" required:"" arg:""`
}

func (cmd *TokenDeleteCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.TokenClient{Client: c}

	t, err := getToken(ctx, c, cmd.Token)
	if err != nil {
		return err
	}

	err = c2.DeleteToken(ctx, t.ID)
	if err != nil {
		return err
	}

	slog.Info("token deleted", slog.Any("id", t.ID))

	return nil
}
