package clients

import (
	"context"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type VolumeProvidersClient struct {
	Client *baseclient.Client
}

func (c *VolumeProvidersClient) CreateVolumeProvider(ctx context.Context, req models.CreateVolumeProvider) (*models.VolumeProvider, error) {
	p, err := c.Client.BuildApiPath(true, "volume-providers")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.VolumeProvider](ctx, c.Client, "POST", p, req)
}

func (c *VolumeProvidersClient) DeleteVolumeProvider(ctx context.Context, volumeProviderId string) error {
	p, err := c.Client.BuildApiPath(true, "volume-providers", volumeProviderId)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[models.Volume](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}

func (c *VolumeProvidersClient) UpdateVolumeProvider(ctx context.Context, volumeProviderId string, req models.UpdateVolumeProvider) (*models.VolumeProvider, error) {
	p, err := c.Client.BuildApiPath(true, "volume-providers", volumeProviderId)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.VolumeProvider](ctx, c.Client, "PATCH", p, req)
}

func (c *VolumeProvidersClient) ListVolumeProviders(ctx context.Context) ([]models.VolumeProvider, error) {
	p, err := c.Client.BuildApiPath(true, "volume-providers")
	if err != nil {
		return nil, err
	}
	l, err := baseclient.RequestApi[huma_utils.ListBody[models.VolumeProvider]](ctx, c.Client, "GET", p, struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, err
}

func (c *VolumeProvidersClient) GetVolumeProviderById(ctx context.Context, volumeProviderId string) (*models.VolumeProvider, error) {
	p, err := c.Client.BuildApiPath(true, "volume-providers", volumeProviderId)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.VolumeProvider](ctx, c.Client, "GET", p, struct{}{})
}

func (c *VolumeProvidersClient) GetVolumeProviderByName(ctx context.Context, name string) (*models.VolumeProvider, error) {
	p, err := c.Client.BuildApiPath(true, "volume-providers", "by-name", name)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.VolumeProvider](ctx, c.Client, "GET", p, struct{}{})
}
