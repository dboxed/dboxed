package clients

import (
	"context"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type GitCredentialsClient struct {
	Client *baseclient.Client
}

func (c *GitCredentialsClient) CreateGitCredentials(ctx context.Context, req models.CreateGitCredentials) (*models.GitCredentials, error) {
	p, err := c.Client.BuildApiPath(true, "git-credentials")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.GitCredentials](ctx, c.Client, "POST", p, req)
}

func (c *GitCredentialsClient) ListGitCredentials(ctx context.Context) ([]models.GitCredentials, error) {
	p, err := c.Client.BuildApiPath(true, "git-credentials")
	if err != nil {
		return nil, err
	}
	l, err := baseclient.RequestApi[huma_utils.ListBody[models.GitCredentials]](ctx, c.Client, "GET", p, struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, err
}

func (c *GitCredentialsClient) GetGitCredentialsById(ctx context.Context, id string) (*models.GitCredentials, error) {
	p, err := c.Client.BuildApiPath(true, "git-credentials", id)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.GitCredentials](ctx, c.Client, "GET", p, struct{}{})
}

func (c *GitCredentialsClient) UpdateGitCredentials(ctx context.Context, id string, req models.UpdateGitCredentials) (*models.GitCredentials, error) {
	p, err := c.Client.BuildApiPath(true, "git-credentials", id)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.GitCredentials](ctx, c.Client, "PATCH", p, req)
}

func (c *GitCredentialsClient) DeleteGitCredentials(ctx context.Context, id string) error {
	p, err := c.Client.BuildApiPath(true, "git-credentials", id)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}
