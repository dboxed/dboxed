package clients

import (
	"context"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type DboxedSpecClient struct {
	Client *baseclient.Client
}

func (c *DboxedSpecClient) CreateDboxedSpec(ctx context.Context, req models.CreateDboxedSpec) (*models.DboxedSpec, error) {
	p, err := c.Client.BuildApiPath(true, "dboxed-specs")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.DboxedSpec](ctx, c.Client, "POST", p, req)
}

func (c *DboxedSpecClient) ListDboxedSpecs(ctx context.Context) ([]models.DboxedSpec, error) {
	p, err := c.Client.BuildApiPath(true, "dboxed-specs")
	if err != nil {
		return nil, err
	}
	l, err := baseclient.RequestApi[huma_utils.ListBody[models.DboxedSpec]](ctx, c.Client, "GET", p, struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, err
}

func (c *DboxedSpecClient) GetDboxedSpecById(ctx context.Context, id string) (*models.DboxedSpec, error) {
	p, err := c.Client.BuildApiPath(true, "dboxed-specs", id)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.DboxedSpec](ctx, c.Client, "GET", p, struct{}{})
}

func (c *DboxedSpecClient) UpdateDboxedSpec(ctx context.Context, id string, req models.UpdateDboxedSpec) (*models.DboxedSpec, error) {
	p, err := c.Client.BuildApiPath(true, "dboxed-specs", id)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.DboxedSpec](ctx, c.Client, "PATCH", p, req)
}

func (c *DboxedSpecClient) DeleteDboxedSpec(ctx context.Context, id string) error {
	p, err := c.Client.BuildApiPath(true, "dboxed-specs", id)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}
