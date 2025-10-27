package clients

import (
	"context"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type S3BucketsClient struct {
	Client *baseclient.Client
}

func (c *S3BucketsClient) CreateS3Bucket(ctx context.Context, req models.CreateS3Bucket) (*models.S3Bucket, error) {
	p, err := c.Client.BuildApiPath(true, "s3-buckets")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.S3Bucket](ctx, c.Client, "POST", p, req)
}

func (c *S3BucketsClient) DeleteS3Bucket(ctx context.Context, s3BucketId int64) error {
	p, err := c.Client.BuildApiPath(true, "s3-buckets", s3BucketId)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}

func (c *S3BucketsClient) ListS3Buckets(ctx context.Context) ([]models.S3Bucket, error) {
	p, err := c.Client.BuildApiPath(true, "s3-buckets")
	if err != nil {
		return nil, err
	}
	l, err := baseclient.RequestApi[huma_utils.ListBody[models.S3Bucket]](ctx, c.Client, "GET", p, struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, err
}

func (c *S3BucketsClient) GetS3BucketById(ctx context.Context, s3BucketId int64) (*models.S3Bucket, error) {
	p, err := c.Client.BuildApiPath(true, "s3-buckets", s3BucketId)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.S3Bucket](ctx, c.Client, "GET", p, struct{}{})
}

func (c *S3BucketsClient) GetS3BucketByBucketName(ctx context.Context, bucket string) (*models.S3Bucket, error) {
	p, err := c.Client.BuildApiPath(true, "s3-buckets", "by-bucket-name", bucket)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.S3Bucket](ctx, c.Client, "GET", p, struct{}{})
}

func (c *S3BucketsClient) UpdateS3Bucket(ctx context.Context, s3BucketId int64, req models.UpdateS3Bucket) (*models.S3Bucket, error) {
	p, err := c.Client.BuildApiPath(true, "s3-buckets", s3BucketId)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.S3Bucket](ctx, c.Client, "PATCH", p, req)
}
