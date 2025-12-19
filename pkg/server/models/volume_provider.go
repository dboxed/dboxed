package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type VolumeProvider struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Workspace string    `json:"workspace"`

	Status        string `json:"status"`
	StatusDetails string `json:"statusDetails"`

	Type dmodel.VolumeProviderType `json:"type"`
	Name string                    `json:"name"`

	Restic *VolumeProviderRestic `json:"restic,omitempty"`
}

type VolumeProviderRestic struct {
	StorageType dmodel.VolumeProviderStorageType `json:"storageType"`
	S3BucketId  *string                          `json:"s3BucketId"`

	StoragePrefix string `json:"storagePrefix"`
}

type CreateVolumeProvider struct {
	Type dmodel.VolumeProviderType `json:"type"`
	Name string                    `json:"name"`

	Restic *CreateVolumeProviderRestic `json:"restic"`
}

type CreateVolumeProviderRestic struct {
	Password string `json:"password"`

	StorageType dmodel.VolumeProviderStorageType `json:"storageType"`
	S3BucketId  *string                          `json:"s3BucketId"`

	StoragePrefix string `json:"storagePrefix"`
}

type UpdateVolumeProvider struct {
	Restic *UpdateVolumeProviderRestic `json:"restic,omitempty"`
}
type UpdateVolumeProviderRestic struct {
	Password *string `json:"password,omitempty"`

	StorageS3 *UpdateRepositoryStorageS3 `json:"storageS3,omitempty"`
}

type UpdateRepositoryStorageS3 struct {
	S3BucketId    *string `json:"s3BucketId"`
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
	if v.Restic != nil && v.Restic.ID.Valid {
		ret.Restic = &VolumeProviderRestic{
			StorageType:   v.Restic.StorageType.V,
			S3BucketId:    v.Restic.S3BucketID,
			StoragePrefix: v.Restic.StoragePrefix.V,
		}
	}
	return ret
}
