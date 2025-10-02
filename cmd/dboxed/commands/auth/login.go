package auth

import (
	"context"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/baseclient"
)

type LoginCmd struct {
}

func (cmd *LoginCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	clientAuth := &baseclient.ClientAuth{}

	if g.ApiUrl != nil {
		clientAuth.ApiUrl = *g.ApiUrl
	}

	c, err := baseclient.New(g.ClientAuthFile, clientAuth, false)
	if err != nil {
		return err
	}

	if g.Workspace != nil {
		w, err := commandutils.GetWorkspace(ctx, c, *g.Workspace)
		if err != nil {
			return err
		}
		_, err = c.SwitchWorkspaceById(ctx, w.ID)
		if err != nil {
			return err
		}
	}

	if g.ApiToken != nil {
		err = c.LoginStaticToken(ctx, *g.ApiToken)
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

	err = c.WriteClientAuth()
	if err != nil {
		return err
	}

	return nil
}
