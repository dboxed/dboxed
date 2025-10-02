package auth

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type TokenCreateCmd struct {
	Name string `help:"Specify the token name. Must be unique." required:"" arg:""`

	ForWorkspace bool    `help:"If set, the token will be for the whole workspace" xor:"for"`
	Box          *string `help:"Specify box for which to create the token" xor:"for"`
}

func (cmd *TokenCreateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.TokenClient{Client: c}

	req := models.CreateToken{
		Name: cmd.Name,
	}

	if cmd.ForWorkspace {
		req.ForWorkspace = true
	} else if cmd.Box != nil {
		b, err := commandutils.GetBox(ctx, c, *cmd.Box)
		if err != nil {
			return err
		}
		req.BoxID = &b.ID
	} else {
		return fmt.Errorf("did not specify for what the token should be")
	}

	token, err := c2.CreateToken(ctx, req)
	if err != nil {
		return err
	}

	slog.Info("token created", slog.Any("id", token.ID), slog.Any("name", token.Name), slog.Any("token", *token.Token))

	return nil
}
