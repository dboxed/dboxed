package commands

import (
	"context"
	"strconv"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type VolumeCmd struct {
	CreateProvider VolumeProviderCreateCmd `cmd:"" help:"Create a provider"`
	UpdateProvider VolumeProviderUpdateCmd `cmd:"" help:"Update a provider"`
	ListProviders  VolumeProviderListCmd   `cmd:"" help:"List providers"`

	Create VolumeCreateCmd `cmd:"" help:"Create a volume"`
	List   VolumeListCmd   `cmd:"" help:"List volumes"`

	VolumeOnlyLinuxCmds
}

func getVolumeProvider(ctx context.Context, c *baseclient.Client, volumeProvider string) (*models.VolumeProvider, error) {
	c2 := clients.VolumeProvidersClient{Client: c}
	id, err := strconv.ParseInt(volumeProvider, 10, 64)
	if err == nil {
		v, err := c2.GetVolumeProviderById(ctx, id)
		if err != nil {
			return nil, err
		}
		return v, nil
	} else {
		v, err := c2.GetVolumeProviderByName(ctx, volumeProvider)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func getVolume(ctx context.Context, c *baseclient.Client, volume string) (*models.Volume, error) {
	c2 := clients.VolumesClient{Client: c}
	id, err := strconv.ParseInt(volume, 10, 64)
	if err == nil {
		v, err := c2.GetVolumeById(ctx, id)
		if err != nil {
			return nil, err
		}
		return v, nil
	} else {
		v, err := c2.GetVolumeByName(ctx, volume)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}
