package commandutils

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type clientTool struct {
	client *baseclient.Client

	Workspaces      cache[models.Workspace]
	Networks        cache[models.Network]
	VolumeProviders cache[models.VolumeProvider]
	Boxes           cache[models.Box]
	S3Buckets       cache[models.S3Bucket]
	Volumes         cache[models.Volume]
	LoadBalancers   cache[models.LoadBalancer]
}

func NewClientTool(c *baseclient.Client) *clientTool {
	ct := &clientTool{
		client: c,
	}

	ct.Workspaces = cache[models.Workspace]{
		getById: func(ctx context.Context, id string) (*models.Workspace, error) {
			return (&clients.WorkspacesClient{Client: c}).GetWorkspaceById(ctx, id)
		},
		entityName: "workspace",
	}
	ct.Networks = cache[models.Network]{
		getById: func(ctx context.Context, id string) (*models.Network, error) {
			return (&clients.NetworkClient{Client: c}).GetNetworkById(ctx, id)
		},
		entityName: "network",
	}
	ct.VolumeProviders = cache[models.VolumeProvider]{
		getById: func(ctx context.Context, id string) (*models.VolumeProvider, error) {
			return (&clients.VolumeProvidersClient{Client: c}).GetVolumeProviderById(ctx, id)
		},
		entityName: "volume provider",
	}
	ct.Boxes = cache[models.Box]{
		getById: func(ctx context.Context, id string) (*models.Box, error) {
			return (&clients.BoxClient{Client: c}).GetBoxById(ctx, id)
		},
		entityName: "box",
	}
	ct.S3Buckets = cache[models.S3Bucket]{
		getById: func(ctx context.Context, id string) (*models.S3Bucket, error) {
			return (&clients.S3BucketsClient{Client: c}).GetS3BucketById(ctx, id)
		},
		entityName: "s3bucket",
		nameField:  "Bucket",
	}
	ct.Volumes = cache[models.Volume]{
		getById: func(ctx context.Context, id string) (*models.Volume, error) {
			return (&clients.VolumesClient{Client: c}).GetVolumeById(ctx, id)
		},
		entityName: "volume",
	}
	ct.LoadBalancers = cache[models.LoadBalancer]{
		getById: func(ctx context.Context, id string) (*models.LoadBalancer, error) {
			return (&clients.LoadBalancerClient{Client: c}).GetLoadBalancerById(ctx, id)
		},
		entityName: "load balancer",
	}

	return ct
}

type cache[T any] struct {
	cache      map[string]*T
	entityName string
	nameField  string
	getById    func(ctx context.Context, id string) (*T, error)
}

func (c *cache[T]) GetColumn(ctx context.Context, id string, showIds bool) string {
	if showIds {
		return id
	}

	if c.cache == nil {
		c.cache = map[string]*T{}
	}
	v, ok := (c.cache)[id]
	if !ok {
		var err error
		v, err = c.getById(ctx, id)
		c.cache[id] = v
		if err != nil {
			slog.WarnContext(ctx, fmt.Sprintf("failed to retrieve %s", c.entityName), slog.Any("error", err))
		}
	}

	var ret string
	if v != nil {
		vv := reflect.Indirect(reflect.ValueOf(v))
		n := c.nameField
		if n == "" {
			n = "Name"
		}
		nameField := vv.FieldByName(n)
		name := nameField.String()
		ret = name
	} else {
		ret = fmt.Sprintf("%s (not found)", id)
	}

	return ret
}
