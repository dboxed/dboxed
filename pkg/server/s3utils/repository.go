package s3utils

import (
	"context"
	"net/url"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func BuildS3ClientFromId(ctx context.Context, bucketId string) (*dmodel.S3Bucket, *minio.Client, error) {
	q := querier.GetQuerier(ctx)
	b, err := dmodel.GetS3BucketById(q, nil, bucketId, true)
	if err != nil {
		return nil, nil, err
	}
	c, err := BuildS3Client(b)
	if err != nil {
		return nil, nil, err
	}
	return b, c, nil
}

func BuildS3Client(b *dmodel.S3Bucket) (*minio.Client, error) {
	creds := credentials.NewStaticV4(b.AccessKeyId, b.SecretAccessKey, "")

	u, err := url.Parse(b.Endpoint)
	if err != nil {
		return nil, err
	}
	region := ""
	if b.DeterminedRegion != nil {
		region = *b.DeterminedRegion
	}
	mc, err := minio.New(u.Host, &minio.Options{
		Creds:  creds,
		Region: region,
		Secure: u.Scheme == "https",
	})
	if err != nil {
		return nil, err
	}

	return mc, nil
}
