package clients

import (
	"context"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type NetworkClient struct {
	Client *baseclient.Client
}

func (c *NetworkClient) ListNetworks(ctx context.Context) ([]models.Network, error) {
	p, err := c.Client.BuildApiPath(true, "networks")
	if err != nil {
		return nil, err
	}
	l, err := baseclient.RequestApi[huma_utils.ListBody[models.Network]](ctx, c.Client, "GET", p, struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, err
}

func (c *NetworkClient) GetNetworkById(ctx context.Context, id int64) (*models.Network, error) {
	p, err := c.Client.BuildApiPath(true, "networks", id)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Network](ctx, c.Client, "GET", p, struct{}{})
}

func (c *NetworkClient) GetNetworkByName(ctx context.Context, name string) (*models.Network, error) {
	p, err := c.Client.BuildApiPath(true, "networks", "by-name", name)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Network](ctx, c.Client, "GET", p, struct{}{})
}

func (c *NetworkClient) CreateNetwork(ctx context.Context, req models.CreateNetwork) (*models.Network, error) {
	p, err := c.Client.BuildApiPath(true, "networks")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Network](ctx, c.Client, "POST", p, req)
}

func (c *NetworkClient) UpdateNetwork(ctx context.Context, id int64, req models.UpdateNetwork) (*models.Network, error) {
	p, err := c.Client.BuildApiPath(true, "networks", id)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Network](ctx, c.Client, "PATCH", p, req)
}

func (c *NetworkClient) DeleteNetwork(ctx context.Context, id int64) error {
	p, err := c.Client.BuildApiPath(true, "networks", id)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[any](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}
