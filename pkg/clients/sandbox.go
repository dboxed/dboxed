package clients

import (
	"context"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type SandboxClient struct {
	Client *baseclient.Client
}

func (c *SandboxClient) ListSandboxes(ctx context.Context) ([]models.BoxSandbox, error) {
	p, err := c.Client.BuildApiPath(true, "sandboxes")
	if err != nil {
		return nil, err
	}
	l, err := baseclient.RequestApi[huma_utils.ListBody[models.BoxSandbox]](ctx, c.Client, "GET", p, struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, nil
}

func (c *SandboxClient) GetSandboxById(ctx context.Context, id string) (*models.BoxSandbox, error) {
	p, err := c.Client.BuildApiPath(true, "sandboxes", id)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.BoxSandbox](ctx, c.Client, "GET", p, struct{}{})
}
