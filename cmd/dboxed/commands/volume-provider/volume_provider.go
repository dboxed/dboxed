package volume_provider

import (
	"context"
	"strconv"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type VolumeProviderCommands struct {
	Create CreateCmd `cmd:"" help:"Create a provider"`
	Update UpdateCmd `cmd:"" help:"Update a provider"`
	Delete DeleteCmd `cmd:"" help:"Delete a provider"`
	List   ListCmd   `cmd:"" help:"List providers"`
}

func GetVolumeProvider(ctx context.Context, c *baseclient.Client, volumeProvider string) (*models.VolumeProvider, error) {
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
