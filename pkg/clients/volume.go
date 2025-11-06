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

func (c *VolumesClient) DeleteVolume(ctx context.Context, volumeId string) error {
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

func (c *VolumesClient) GetVolumeById(ctx context.Context, volumeId string) (*models.Volume, error) {
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

func (c *VolumesClient) VolumeLock(ctx context.Context, volumeId string, req models.VolumeLockRequest) (*models.Volume, error) {
	p, err := c.Client.BuildApiPath(true, "volumes", volumeId, "lock")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Volume](ctx, c.Client, "POST", p, req)
}

func (c *VolumesClient) VolumeRefreshLock(ctx context.Context, volumeId string, req models.VolumeRefreshLockRequest) (*models.Volume, error) {
	p, err := c.Client.BuildApiPath(true, "volumes", volumeId, "refresh-lock")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Volume](ctx, c.Client, "POST", p, req)
}

func (c *VolumesClient) VolumeRelease(ctx context.Context, volumeId string, req models.VolumeReleaseRequest) (*models.Volume, error) {
	p, err := c.Client.BuildApiPath(true, "volumes", volumeId, "release")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Volume](ctx, c.Client, "POST", p, req)
}

func (c *VolumesClient) VolumeForceUnlock(ctx context.Context, volumeId string) (*models.Volume, error) {
	p, err := c.Client.BuildApiPath(true, "volumes", volumeId, "force-unlock")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Volume](ctx, c.Client, "POST", p, struct{}{})
}

func (c *VolumesClient) CreateSnapshot(ctx context.Context, volumeId string, req models.CreateVolumeSnapshot) (*models.VolumeSnapshot, error) {
	p, err := c.Client.BuildApiPath(true, "volumes", volumeId, "snapshots")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.VolumeSnapshot](ctx, c.Client, "POST", p, req)
}

func (c *VolumesClient) GetVolumeSnapshotById(ctx context.Context, volumeId string, snapshotId string) (*models.VolumeSnapshot, error) {
	p, err := c.Client.BuildApiPath(true, "volumes", volumeId, "snapshots", snapshotId)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.VolumeSnapshot](ctx, c.Client, "GET", p, struct{}{})
}
