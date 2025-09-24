package clients

import (
	"context"
	"fmt"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type TokenClient struct {
	Client *baseclient.Client
}

func (c *TokenClient) CreateToken(ctx context.Context, req models.CreateToken) (*models.CreateTokenResult, error) {
	return baseclient.RequestApi[models.CreateTokenResult](ctx, c.Client, "POST", "v1/tokens", req)
}

func (c *TokenClient) ListTokens(ctx context.Context) ([]models.Token, error) {
	l, err := baseclient.RequestApi[huma_utils.ListBody[models.Token]](ctx, c.Client, "GET", "v1/tokens", struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, err
}

func (c *TokenClient) GetTokenById(ctx context.Context, tokenId string) (*models.Token, error) {
	return baseclient.RequestApi[models.Token](ctx, c.Client, "GET", fmt.Sprintf("v1/tokens/%s", tokenId), struct{}{})
}

func (c *TokenClient) DeleteToken(ctx context.Context, tokenId string) error {
	_, err := baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "DELETE", fmt.Sprintf("v1/tokens/%s", tokenId), struct{}{})
	return err
}
