package dmodel

import (
	"github.com/dboxed/dboxed/pkg/server/db/querier"
)

type VolumeProvider struct {
	OwnedByWorkspace
	ReconcileStatus

	Type string `db:"type"`
	Name string `db:"name"`

	Rustic *VolumeProviderRustic `join:"true"`
}

type VolumeProviderRustic struct {
	ID querier.NullForJoin[int64] `db:"id"`

	Password querier.NullForJoin[string] `db:"password"`

	StorageType string `db:"storage_type"`

	StorageS3 *VolumeProviderStorageS3 `join:"true" db:"storage_s3"`
}

type VolumeProviderStorageS3 struct {
	ID querier.NullForJoin[int64] `db:"id"`

	Endpoint        querier.NullForJoin[string] `db:"endpoint"`
	Region          *string                     `db:"region"`
	Bucket          querier.NullForJoin[string] `db:"bucket"`
	Prefix          querier.NullForJoin[string] `db:"prefix"`
	AccessKeyId     querier.NullForJoin[string] `db:"access_key_id"`
	SecretAccessKey querier.NullForJoin[string] `db:"secret_access_key"`
}

func postprocessVolumeProvider(q *querier.Querier, vr *VolumeProvider) error {
	return nil
}

func (v *VolumeProvider) Create(q *querier.Querier) error {
	return querier.Create(q, v)
}

func (v *VolumeProviderRustic) Create(q *querier.Querier) error {
	return querier.Create(q, v)
}

func (v *VolumeProviderStorageS3) Create(q *querier.Querier) error {
	return querier.Create(q, v)
}

func ListVolumeProviders(q *querier.Querier, workspaceId *int64, skipDeleted bool) ([]VolumeProvider, error) {
	l, err := querier.GetMany[VolumeProvider](q, map[string]any{
		"workspace_id": querier.OmitIfNull(workspaceId),
		"deleted_at":   querier.ExcludeNonNull(skipDeleted),
	})
	if err != nil {
		return nil, err
	}

	var ret []VolumeProvider
	for _, n := range l {
		err = postprocessVolumeProvider(q, &n)
		if err != nil {
			return nil, err
		}
		ret = append(ret, n)
	}
	return ret, nil
}

func GetVolumeProviderById(q *querier.Querier, workspaceId *int64, id int64, skipDeleted bool) (*VolumeProvider, error) {
	vp, err := querier.GetOne[VolumeProvider](q, map[string]any{
		"workspace_id": querier.OmitIfNull(workspaceId),
		"id":           id,
		"deleted_at":   querier.ExcludeNonNull(skipDeleted),
	})
	if err != nil {
		return nil, err
	}
	err = postprocessVolumeProvider(q, vp)
	if err != nil {
		return nil, err
	}
	return vp, nil
}

func GetVolumeProviderByName(q *querier.Querier, workspaceId int64, name string, skipDeleted bool) (*VolumeProvider, error) {
	vp, err := querier.GetOne[VolumeProvider](q, map[string]any{
		"workspace_id": workspaceId,
		"name":         name,
		"deleted_at":   querier.ExcludeNonNull(skipDeleted),
	})
	if err != nil {
		return nil, err
	}
	err = postprocessVolumeProvider(q, vp)
	if err != nil {
		return nil, err
	}
	return vp, nil
}

func (v *VolumeProviderRustic) UpdatePassword(q *querier.Querier, password string) error {
	v.Password = querier.N(password)
	return querier.UpdateOneFromStruct(q, v,
		"password",
	)
}

func (v *VolumeProviderStorageS3) UpdateEndpoint(q *querier.Querier, endpoint string) error {
	v.Endpoint = querier.N(endpoint)
	return querier.UpdateOneFromStruct(q, v,
		"endpoint",
	)
}

func (v *VolumeProviderStorageS3) UpdateRegion(q *querier.Querier, region *string) error {
	v.Region = region
	return querier.UpdateOneFromStruct(q, v,
		"region",
	)
}

func (v *VolumeProviderStorageS3) UpdateBucket(q *querier.Querier, bucket string) error {
	v.Bucket = querier.N(bucket)
	return querier.UpdateOneFromStruct(q, v,
		"bucket",
	)
}

func (v *VolumeProviderStorageS3) UpdatePrefix(q *querier.Querier, prefix string) error {
	v.Prefix = querier.N(prefix)
	return querier.UpdateOneFromStruct(q, v,
		"prefix",
	)
}

func (v *VolumeProviderStorageS3) UpdateKeys(q *querier.Querier, accessKeyId string, secretAccessKey string) error {
	v.AccessKeyId = querier.N(accessKeyId)
	v.SecretAccessKey = querier.N(secretAccessKey)
	return querier.UpdateOneFromStruct(q, v,
		"access_key_id",
		"secret_access_key",
	)
}
