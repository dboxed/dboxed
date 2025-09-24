package auth

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
)

type TokenDeleteCmd struct {
	Id string `help:"Token ID" required:"" arg:""`
}

func (cmd *TokenDeleteCmd) Run() error {
	ctx := context.Background()

	c, err := baseclient.FromClientAuthFile()
	if err != nil {
		return err
	}

	c2 := &clients.TokenClient{Client: c}

	err = c2.DeleteToken(ctx, cmd.Id)
	if err != nil {
		return err
	}

	slog.Info("token deleted", slog.Any("id", cmd.Id))

	return nil
}
