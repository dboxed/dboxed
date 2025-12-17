package git_credentials

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type UpdateCmd struct {
	GitCredentials string `help:"Specify git credentials ID or host" required:"" arg:""`

	Username *string `help:"Username for basic auth"`
	Password *string `help:"Password for basic auth"`
	SshKey   *string `help:"SSH private key"`
}

func (cmd *UpdateCmd) Run(g *flags.GlobalFlags) error {
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

	req := models.UpdateGitCredentials{
		Username: cmd.Username,
		Password: cmd.Password,
		SshKey:   cmd.SshKey,
	}

	updated, err := c2.UpdateGitCredentials(ctx, gc.ID, req)
	if err != nil {
		return err
	}

	slog.Info("git credentials updated", slog.Any("id", updated.ID), slog.Any("host", updated.Host))

	return nil
}
