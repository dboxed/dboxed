package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type VolumeProvider struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Workspace int64     `json:"workspace"`

	Status        string `json:"status"`
	StatusDetails string `json:"statusDetails"`

	Type dmodel.VolumeProviderType `json:"type"`
	Name string                    `json:"name"`

	Rustic *VolumeProviderRustic `json:"rustic,omitempty"`
}

type VolumeProviderRustic struct {
	StorageType dmodel.VolumeProviderStorageType `json:"storageType"`
	S3BucketId  *int64                           `json:"s3BucketId"`

	StoragePrefix string `json:"storagePrefix"`
}

type CreateVolumeProvider struct {
	Type dmodel.VolumeProviderType `json:"type"`
	Name string                    `json:"name"`

	Rustic *CreateVolumeProviderRustic `json:"rustic"`
}

type CreateVolumeProviderRustic struct {
	Password string `json:"password"`

	StorageType dmodel.VolumeProviderStorageType `json:"storageType"`
	S3BucketId  *int64                           `json:"s3BucketId"`

	StoragePrefix string `json:"storagePrefix"`
}

type UpdateVolumeProvider struct {
	Rustic *UpdateVolumeProviderRustic `json:"rustic,omitempty"`
}
type UpdateVolumeProviderRustic struct {
	Password *string `json:"password,omitempty"`

	StorageS3 *UpdateRepositoryStorageS3 `json:"storageS3,omitempty"`
}

type UpdateRepositoryStorageS3 struct {
	S3BucketId    *int64  `json:"s3BucketId"`
	StoragePrefix *string `json:"storagePrefix"`
}

func VolumeProviderFromDB(v dmodel.VolumeProvider) VolumeProvider {
	ret := VolumeProvider{
		ID:            v.ID,
		CreatedAt:     v.CreatedAt,
		Workspace:     v.WorkspaceID,
		Status:        v.ReconcileStatus.ReconcileStatus.V,
		StatusDetails: v.ReconcileStatus.ReconcileStatusDetails.V,

		Type: v.Type,
		Name: v.Name,
	}
	if v.Rustic != nil && v.Rustic.ID.Valid {
		ret.Rustic = &VolumeProviderRustic{
			StorageType:   v.Rustic.StorageType.V,
			S3BucketId:    v.Rustic.S3BucketID,
			StoragePrefix: v.Rustic.StoragePrefix.V,
		}
	}
	return ret
}
