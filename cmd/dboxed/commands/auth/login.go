package auth

import (
	"context"

	"github.com/dboxed/dboxed/pkg/baseclient"
)

type LoginCmd struct {
	ApiUrl   *string `help:"Specify the API url"`
	ApiToken *string `help:"Specify a static API token"`
}

func (cmd *LoginCmd) Run() error {
	ctx := context.Background()

	c, err := baseclient.New(cmd.ApiUrl, true)
	if err != nil {
		return err
	}

	if cmd.ApiToken != nil {
		err = c.LoginStaticToken(ctx, *cmd.ApiToken)
		if err != nil {
			return err
		}
	} else {
		err = c.LoginOAuth2(ctx)
		if err != nil {
			return err
		}
	}

	err = c.CheckAuth(ctx)
	if err != nil {
		return err
	}

	return nil
}
