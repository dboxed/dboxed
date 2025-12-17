package git_credentials

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type CreateCmd struct {
	Host     string `help:"Git host (e.g., github.com)" required:""`
	PathGlob string `help:"Path glob for matching repositories (e.g., myorg/*)"`

	Username *string `help:"Username for basic auth"`
	Password *string `help:"Password for basic auth"`

	SshKey *string `help:"SSH private key"`
}

func (cmd *CreateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	// Auto-detect credentials type
	var credentialsType dmodel.GitCredentialsType
	if cmd.SshKey != nil {
		credentialsType = dmodel.GitCredentialsTypeSshKey
	} else {
		credentialsType = dmodel.GitCredentialsTypeBasicAuth
	}

	c2 := &clients.GitCredentialsClient{Client: c}

	req := models.CreateGitCredentials{
		Host:            cmd.Host,
		PathGlob:        cmd.PathGlob,
		CredentialsType: credentialsType,
		Username:        cmd.Username,
		Password:        cmd.Password,
		SshKey:          cmd.SshKey,
	}

	gc, err := c2.CreateGitCredentials(ctx, req)
	if err != nil {
		return err
	}

	slog.Info("git credentials created", slog.Any("id", gc.ID), slog.Any("host", gc.Host))

	return nil
}
