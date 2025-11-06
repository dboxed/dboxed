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

	return ct
}

type cache[T any] struct {
	cache      map[string]*T
	entityName string
	nameField  string
	getById    func(ctx context.Context, id string) (*T, error)
}

func (c *cache[T]) GetColumn(ctx context.Context, id string) string {
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
		ret = fmt.Sprintf("%s (id=%s)", name, id)
	} else {
		ret = fmt.Sprintf("<unknown> (id=%s)", id)
	}

	return ret
}
