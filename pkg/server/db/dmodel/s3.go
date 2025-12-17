package dmodel

import "github.com/dboxed/dboxed/pkg/server/db/querier"

type S3Bucket struct {
	OwnedByWorkspace
	SoftDeleteFields
	ReconcileStatus

	Endpoint        string `db:"endpoint"`
	Bucket          string `db:"bucket"`
	AccessKeyId     string `db:"access_key_id"`
	SecretAccessKey string `db:"secret_access_key"`
}

func (v *S3Bucket) Create(q *querier.Querier) error {
	return querier.Create(q, v)
}

func ListS3Buckets(q *querier.Querier, workspaceId *string, skipDeleted bool) ([]S3Bucket, error) {
	l, err := querier.GetMany[S3Bucket](q, map[string]any{
		"workspace_id": querier.OmitIfNull(workspaceId),
		"deleted_at":   querier.ExcludeNonNull(skipDeleted),
	}, nil)
	if err != nil {
		return nil, err
	}

	return l, nil
}

func GetS3BucketById(q *querier.Querier, workspaceId *string, id string, skipDeleted bool) (*S3Bucket, error) {
	vp, err := querier.GetOne[S3Bucket](q, map[string]any{
		"workspace_id": querier.OmitIfNull(workspaceId),
		"id":           id,
		"deleted_at":   querier.ExcludeNonNull(skipDeleted),
	})
	if err != nil {
		return nil, err
	}
	return vp, nil
}

func GetS3BucketByBucketName(q *querier.Querier, workspaceId string, bucket string, skipDeleted bool) (*S3Bucket, error) {
	return querier.GetOne[S3Bucket](q, map[string]any{
		"workspace_id": workspaceId,
		"bucket":       bucket,
		"deleted_at":   querier.ExcludeNonNull(skipDeleted),
	})
}

func (v *S3Bucket) UpdateEndpoint(q *querier.Querier, endpoint string) error {
	v.Endpoint = endpoint
	return querier.UpdateOneFromStruct(q, v,
		"endpoint",
	)
}

func (v *S3Bucket) UpdateBucket(q *querier.Querier, bucket string) error {
	v.Bucket = bucket
	return querier.UpdateOneFromStruct(q, v,
		"bucket",
	)
}

func (v *S3Bucket) UpdateKeys(q *querier.Querier, accessKeyId string, secretAccessKey string) error {
	v.AccessKeyId = accessKeyId
	v.SecretAccessKey = secretAccessKey
	return querier.UpdateOneFromStruct(q, v,
		"access_key_id",
		"secret_access_key",
	)
}
