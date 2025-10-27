package dmodel

import (
	"github.com/dboxed/dboxed/pkg/server/db/querier"
)

type VolumeProvider struct {
	OwnedByWorkspace
	ReconcileStatus

	Type VolumeProviderType `db:"type"`
	Name string             `db:"name"`

	Rustic *VolumeProviderRustic `join:"true"`
}

type VolumeProviderRustic struct {
	ID querier.NullForJoin[int64] `db:"id"`

	Password querier.NullForJoin[string] `db:"password"`

	StorageType querier.NullForJoin[VolumeProviderStorageType] `db:"storage_type"`

	S3BucketID    *int64                      `db:"s3_bucket_id"`
	StoragePrefix querier.NullForJoin[string] `db:"storage_prefix"`
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

func ListVolumeProviders(q *querier.Querier, workspaceId *int64, skipDeleted bool) ([]VolumeProvider, error) {
	l, err := querier.GetMany[VolumeProvider](q, map[string]any{
		"workspace_id": querier.OmitIfNull(workspaceId),
		"deleted_at":   querier.ExcludeNonNull(skipDeleted),
	}, nil)
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

func (v *VolumeProviderRustic) UpdateS3Bucket(q *querier.Querier, bucketId int64) error {
	v.S3BucketID = &bucketId
	return querier.UpdateOneFromStruct(q, v,
		"s3_bucket_id",
	)
}

func (v *VolumeProviderRustic) UpdatePassword(q *querier.Querier, password string) error {
	v.Password = querier.N(password)
	return querier.UpdateOneFromStruct(q, v,
		"password",
	)
}

func (v *VolumeProviderRustic) UpdateStoragePrefix(q *querier.Querier, prefix string) error {
	v.StoragePrefix = querier.N(prefix)
	return querier.UpdateOneFromStruct(q, v,
		"storage_prefix",
	)
}
