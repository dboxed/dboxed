package commandutils

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/google/uuid"
)

func GetWorkspace(ctx context.Context, c *baseclient.Client, workspace string) (*models.Workspace, error) {
	c2 := clients.WorkspacesClient{Client: c}
	id, err := strconv.ParseInt(workspace, 10, 64)
	if err == nil {
		v, err := c2.GetWorkspaceById(ctx, id)
		if err != nil {
			return nil, err
		}
		return v, nil
	} else {
		l, err := c2.ListWorkspaces(ctx)
		if err != nil {
			return nil, err
		}
		for _, w := range l {
			if w.Name == workspace {
				return &w, nil
			}
		}
		return nil, fmt.Errorf("workspace not found")
	}
}

func GetBox(ctx context.Context, c *baseclient.Client, box string) (*models.Box, error) {
	c2 := clients.BoxClient{Client: c}
	id, err := strconv.ParseInt(box, 10, 64)
	if err == nil {
		v, err := c2.GetBoxById(ctx, id)
		if err != nil {
			return nil, err
		}
		return v, nil
	} else {
		err = uuid.Validate(box)
		if err == nil {
			v, err := c2.GetBoxByUuid(ctx, box)
			if err != nil {
				if !baseclient.IsNotFound(err) {
					return nil, err
				}
			} else {
				return v, nil
			}
		}
		v, err := c2.GetBoxByName(ctx, box)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
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

func GetVolume(ctx context.Context, c *baseclient.Client, volume string) (*models.Volume, error) {
	c2 := clients.VolumesClient{Client: c}
	id, err := strconv.ParseInt(volume, 10, 64)
	if err == nil {
		v, err := c2.GetVolumeById(ctx, id)
		if err != nil {
			return nil, err
		}
		return v, nil
	} else {
		err = uuid.Validate(volume)
		if err == nil {
			v, err := c2.GetVolumeByUuid(ctx, volume)
			if err != nil {
				if !baseclient.IsNotFound(err) {
					return nil, err
				}
			} else {
				return v, nil
			}
		}

		v, err := c2.GetVolumeByName(ctx, volume)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func GetNetwork(ctx context.Context, c *baseclient.Client, network string) (*models.Network, error) {
	c2 := clients.NetworkClient{Client: c}
	id, err := strconv.ParseInt(network, 10, 64)
	if err == nil {
		n, err := c2.GetNetworkById(ctx, id)
		if err != nil {
			return nil, err
		}
		return n, nil
	} else {
		n, err := c2.GetNetworkByName(ctx, network)
		if err != nil {
			return nil, err
		}
		return n, nil
	}
}
