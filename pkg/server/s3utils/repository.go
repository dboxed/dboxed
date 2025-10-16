package s3utils

import (
	"fmt"
	"net/url"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func BuildS3Client(vp *dmodel.VolumeProvider, region string) (*minio.Client, error) {
	if vp.Rustic == nil || vp.Rustic.StorageS3 == nil {
		return nil, fmt.Errorf("not a S3 repository")
	}

	creds := credentials.NewStaticV4(vp.Rustic.StorageS3.AccessKeyId.V, vp.Rustic.StorageS3.SecretAccessKey.V, "")

	u, err := url.Parse(vp.Rustic.StorageS3.Endpoint.V)
	if err != nil {
		return nil, err
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
