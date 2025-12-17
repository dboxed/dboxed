package git_credentials

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type DeleteCmd struct {
	GitCredentials string `help:"Specify git credentials ID or host" required:"" arg:""`
}

func (cmd *DeleteCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	gc, err := commandutils.GetGitCredentials(ctx, c, cmd.GitCredentials)
	if err != nil {
		return err
	}

	c2 := &clients.GitCredentialsClient{Client: c}

	err = c2.DeleteGitCredentials(ctx, gc.ID)
	if err != nil {
		return err
	}

	slog.Info("git credentials deleted", slog.Any("id", gc.ID), slog.Any("host", gc.Host))

	return nil
}
