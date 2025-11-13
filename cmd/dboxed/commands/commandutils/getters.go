package commandutils

import (
	"context"
	"fmt"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/google/uuid"
)

func GetWorkspace(ctx context.Context, c *baseclient.Client, workspace string) (*models.Workspace, error) {
	c2 := clients.WorkspacesClient{Client: c}
	if uuid.Validate(workspace) == nil {
		v, err := c2.GetWorkspaceById(ctx, workspace)
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
	if uuid.Validate(box) == nil {
		v, err := c2.GetBoxById(ctx, box)
		if err != nil {
			return nil, err
		}
		return v, nil
	} else {
		v, err := c2.GetBoxByName(ctx, box)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func GetVolumeProvider(ctx context.Context, c *baseclient.Client, volumeProvider string) (*models.VolumeProvider, error) {
	c2 := clients.VolumeProvidersClient{Client: c}
	if uuid.Validate(volumeProvider) == nil {
		v, err := c2.GetVolumeProviderById(ctx, volumeProvider)
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
	if uuid.Validate(volume) == nil {
		v, err := c2.GetVolumeById(ctx, volume)
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

func GetNetwork(ctx context.Context, c *baseclient.Client, network string) (*models.Network, error) {
	c2 := clients.NetworkClient{Client: c}
	if uuid.Validate(network) == nil {
		n, err := c2.GetNetworkById(ctx, network)
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

func GetS3Bucket(ctx context.Context, c *baseclient.Client, s3Bucket string) (*models.S3Bucket, error) {
	c2 := clients.S3BucketsClient{Client: c}
	if uuid.Validate(s3Bucket) == nil {
		s, err := c2.GetS3BucketById(ctx, s3Bucket)
		if err != nil {
			return nil, err
		}
		return s, nil
	} else {
		s, err := c2.GetS3BucketByBucketName(ctx, s3Bucket)
		if err != nil {
			return nil, err
		}
		return s, nil
	}
}

func GetIngressProxy(ctx context.Context, c *baseclient.Client, proxy string) (*models.IngressProxy, error) {
	c2 := clients.IngressProxyClient{Client: c}
	if uuid.Validate(proxy) == nil {
		p, err := c2.GetIngressProxyById(ctx, proxy)
		if err != nil {
			return nil, err
		}
		return p, nil
	} else {
		l, err := c2.ListIngressProxies(ctx)
		if err != nil {
			return nil, err
		}
		for _, p := range l {
			if p.Name == proxy {
				return &p, nil
			}
		}
		return nil, fmt.Errorf("ingress proxy not found: %s", proxy)
	}
}
