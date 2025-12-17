package clients

import (
	"context"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type GitSpecClient struct {
	Client *baseclient.Client
}

func (c *GitSpecClient) CreateGitSpec(ctx context.Context, req models.CreateGitSpec) (*models.GitSpec, error) {
	p, err := c.Client.BuildApiPath(true, "git-specs")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.GitSpec](ctx, c.Client, "POST", p, req)
}

func (c *GitSpecClient) ListGitSpecs(ctx context.Context) ([]models.GitSpec, error) {
	p, err := c.Client.BuildApiPath(true, "git-specs")
	if err != nil {
		return nil, err
	}
	l, err := baseclient.RequestApi[huma_utils.ListBody[models.GitSpec]](ctx, c.Client, "GET", p, struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, err
}

func (c *GitSpecClient) GetGitSpecById(ctx context.Context, id string) (*models.GitSpec, error) {
	p, err := c.Client.BuildApiPath(true, "git-specs", id)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.GitSpec](ctx, c.Client, "GET", p, struct{}{})
}

func (c *GitSpecClient) UpdateGitSpec(ctx context.Context, id string, req models.UpdateGitSpec) (*models.GitSpec, error) {
	p, err := c.Client.BuildApiPath(true, "git-specs", id)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.GitSpec](ctx, c.Client, "PATCH", p, req)
}

func (c *GitSpecClient) DeleteGitSpec(ctx context.Context, id string) error {
	p, err := c.Client.BuildApiPath(true, "git-specs", id)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}
