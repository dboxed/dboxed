package clients

import (
	"context"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type TokenClient struct {
	Client *baseclient.Client
}

func (c *TokenClient) CreateToken(ctx context.Context, req models.CreateToken) (*models.Token, error) {
	p, err := c.Client.BuildApiPath(true, "tokens")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Token](ctx, c.Client, "POST", p, req)
}

func (c *TokenClient) ListTokens(ctx context.Context) ([]models.Token, error) {
	p, err := c.Client.BuildApiPath(true, "tokens")
	if err != nil {
		return nil, err
	}
	l, err := baseclient.RequestApi[huma_utils.ListBody[models.Token]](ctx, c.Client, "GET", p, struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, err
}

func (c *TokenClient) GetTokenById(ctx context.Context, tokenId string) (*models.Token, error) {
	p, err := c.Client.BuildApiPath(true, "tokens", tokenId)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Token](ctx, c.Client, "GET", p, struct{}{})
}

func (c *TokenClient) GetTokenByName(ctx context.Context, name string) (*models.Token, error) {
	p, err := c.Client.BuildApiPath(true, "tokens", "by-name", name)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Token](ctx, c.Client, "GET", p, struct{}{})
}

func (c *TokenClient) DeleteToken(ctx context.Context, tokenId string) error {
	p, err := c.Client.BuildApiPath(true, "tokens", tokenId)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}
