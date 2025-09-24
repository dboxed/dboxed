package clients

import (
	"context"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type VolumesClient struct {
	Client *baseclient.Client
}

func (c *VolumesClient) CreateVolume(ctx context.Context, req models.CreateVolume) (*models.Volume, error) {
	p, err := c.Client.BuildApiPath(true, "volumes")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Volume](ctx, c.Client, "POST", p, req)
}

func (c *VolumesClient) DeleteVolume(ctx context.Context, volumeId int64) error {
	p, err := c.Client.BuildApiPath(true, "volumes", volumeId)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}

func (c *VolumesClient) ListVolumes(ctx context.Context) ([]models.Volume, error) {
	p, err := c.Client.BuildApiPath(true, "volumes")
	if err != nil {
		return nil, err
	}
	l, err := baseclient.RequestApi[huma_utils.ListBody[models.Volume]](ctx, c.Client, "GET", p, struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, err
}

func (c *VolumesClient) GetVolumeById(ctx context.Context, volumeId int64) (*models.Volume, error) {
	p, err := c.Client.BuildApiPath(true, "volumes", volumeId)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Volume](ctx, c.Client, "GET", p, struct{}{})
}

func (c *VolumesClient) GetVolumeByName(ctx context.Context, name string) (*models.Volume, error) {
	p, err := c.Client.BuildApiPath(true, "volumes", "by-name", name)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Volume](ctx, c.Client, "GET", p, struct{}{})
}

func (c *VolumesClient) VolumeLock(ctx context.Context, volumeId int64, req models.VolumeLockRequest) (*models.Volume, error) {
	p, err := c.Client.BuildApiPath(true, "volumes", volumeId, "lock")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Volume](ctx, c.Client, "POST", p, req)
}

func (c *VolumesClient) VolumeRelease(ctx context.Context, volumeId int64, req models.VolumeReleaseRequest) (*models.Volume, error) {
	p, err := c.Client.BuildApiPath(true, "volumes", volumeId, "release")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Volume](ctx, c.Client, "POST", p, req)
}

func (c *VolumesClient) VolumeCreateSnapshot(ctx context.Context, volumeId int64, req models.CreateVolumeSnapshot) (*models.Volume, error) {
	p, err := c.Client.BuildApiPath(true, "volumes", volumeId, "snapshots")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Volume](ctx, c.Client, "POST", p, req)
}

func (c *VolumesClient) GetVolumeSnapshotById(ctx context.Context, volumeId int64, snapshotId int64) (*models.VolumeSnapshot, error) {
	p, err := c.Client.BuildApiPath(true, "volumes", volumeId, "snapshots", snapshotId)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.VolumeSnapshot](ctx, c.Client, "GET", p, struct{}{})
}
