package clients

import (
	"context"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type S3ProxyClient struct {
	Client *baseclient.Client
}

func (c *S3ProxyClient) S3ProxyListObjects(ctx context.Context, s3BucketId int64, req models.S3ProxyListObjectsRequest) (*models.S3ProxyListObjectsResult, error) {
	p, err := c.Client.BuildApiPath(true, "s3-buckets", s3BucketId, "proxy", "list-objects")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.S3ProxyListObjectsResult](ctx, c.Client, "POST", p, req)
}

func (c *S3ProxyClient) S3ProxyPresignPut(ctx context.Context, s3BucketId int64, req models.S3ProxyPresignPutRequest) (*models.S3ProxyPresignPutResult, error) {
	p, err := c.Client.BuildApiPath(true, "s3-buckets", s3BucketId, "proxy", "presign-put")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.S3ProxyPresignPutResult](ctx, c.Client, "POST", p, req)
}

func (c *S3ProxyClient) S3ProxyRenameObject(ctx context.Context, s3BucketId int64, req models.S3ProxyRenameObjectRequest) (*models.S3ProxyRenameObjectResult, error) {
	p, err := c.Client.BuildApiPath(true, "s3-buckets", s3BucketId, "proxy", "rename-object")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.S3ProxyRenameObjectResult](ctx, c.Client, "POST", p, req)
}

func (c *S3ProxyClient) S3ProxyDeleteObject(ctx context.Context, s3BucketId int64, req models.S3ProxyDeleteObjectRequest) (*models.S3ProxyDeleteObjectResult, error) {
	p, err := c.Client.BuildApiPath(true, "s3-buckets", s3BucketId, "proxy", "delete-object")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.S3ProxyDeleteObjectResult](ctx, c.Client, "POST", p, req)
}
