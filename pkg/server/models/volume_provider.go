package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type VolumeProvider struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Workspace int64     `json:"workspace"`
	Status    string    `json:"status"`

	Type dmodel.VolumeProviderType `json:"type"`
	Name string                    `json:"name"`

	Rustic *VolumeProviderRustic `json:"rustic,omitempty"`
}

type VolumeProviderRustic struct {
	StorageType dmodel.VolumeProviderStorageType `json:"storageType"`
	StorageS3   *VolumeStorageS3                 `json:"storageS3"`
}

type VolumeStorageS3 struct {
	Endpoint string `json:"endpoint"`
	Bucket   string `json:"bucket"`
	Prefix   string `json:"prefix"`
}

type CreateVolumeProvider struct {
	Type dmodel.VolumeProviderType `json:"type"`
	Name string                    `json:"name"`

	Rustic *CreateVolumeProviderRustic `json:"rustic"`
}

type CreateVolumeProviderStorageS3 struct {
	Endpoint        string `json:"endpoint"`
	Bucket          string `json:"bucket"`
	Prefix          string `json:"prefix"`
	AccessKeyId     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
}

type CreateVolumeProviderRustic struct {
	Password string `json:"password"`

	StorageType dmodel.VolumeProviderStorageType `json:"storageType"`
	StorageS3   *CreateVolumeProviderStorageS3   `json:"storageS3"`
}

type UpdateVolumeProvider struct {
	Rustic *UpdateVolumeProviderRustic `json:"rustic,omitempty"`
}
type UpdateVolumeProviderRustic struct {
	Password *string `json:"password,omitempty"`

	StorageS3 *UpdateRepositoryStorageS3 `json:"storageS3,omitempty"`
}

type UpdateRepositoryStorageS3 struct {
	Endpoint        *string `json:"endpoint,omitempty"`
	Bucket          *string `json:"bucket,omitempty"`
	Prefix          *string `json:"prefix,omitempty"`
	AccessKeyId     *string `json:"accessKeyId,omitempty"`
	SecretAccessKey *string `json:"secretAccessKey,omitempty"`
}

func VolumeProviderFromDB(v dmodel.VolumeProvider) VolumeProvider {
	ret := VolumeProvider{
		ID:        v.ID,
		CreatedAt: v.CreatedAt,
		Workspace: v.WorkspaceID,
		Type:      dmodel.VolumeProviderType(v.Type),
		Name:      v.Name,
		Status:    v.ReconcileStatus.ReconcileStatus,
	}
	if v.Rustic != nil && v.Rustic.ID.Valid {
		ret.Rustic = &VolumeProviderRustic{
			StorageType: v.Rustic.StorageType,
		}
		if v.Rustic.StorageS3 != nil && v.Rustic.StorageS3.ID.Valid {
			ret.Rustic.StorageS3 = &VolumeStorageS3{
				Endpoint: v.Rustic.StorageS3.Endpoint.V,
				Bucket:   v.Rustic.StorageS3.Bucket.V,
				Prefix:   v.Rustic.StorageS3.Prefix.V,
			}
		}
	}
	return ret
}
