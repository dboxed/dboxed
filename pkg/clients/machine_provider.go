package clients

import (
	"context"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type MachineProviderClient struct {
	Client *baseclient.Client
}

func (c *MachineProviderClient) CreateMachineProvider(ctx context.Context, req models.CreateMachineProvider) (*models.MachineProvider, error) {
	p, err := c.Client.BuildApiPath(true, "machine-providers")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.MachineProvider](ctx, c.Client, "POST", p, req)
}

func (c *MachineProviderClient) ListMachineProviders(ctx context.Context) ([]models.MachineProvider, error) {
	p, err := c.Client.BuildApiPath(true, "machine-providers")
	if err != nil {
		return nil, err
	}
	l, err := baseclient.RequestApi[huma_utils.ListBody[models.MachineProvider]](ctx, c.Client, "GET", p, struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, err
}

func (c *MachineProviderClient) GetMachineProviderById(ctx context.Context, id string) (*models.MachineProvider, error) {
	p, err := c.Client.BuildApiPath(true, "machine-providers", id)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.MachineProvider](ctx, c.Client, "GET", p, struct{}{})
}

func (c *MachineProviderClient) DeleteMachineProvider(ctx context.Context, id string) error {
	p, err := c.Client.BuildApiPath(true, "machine-providers", id)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}

func (c *MachineProviderClient) UpdateMachineProvider(ctx context.Context, id string, req models.UpdateMachineProvider) (*models.MachineProvider, error) {
	p, err := c.Client.BuildApiPath(true, "machine-providers", id)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.MachineProvider](ctx, c.Client, "PATCH", p, req)
}
