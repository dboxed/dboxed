package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/global"
)

type VolumeProvider struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Workspace int64     `json:"workspace"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`

	Dboxed *VolumeProviderDboxed `json:"dboxed,omitempty"`
}

type VolumeProviderDboxed struct {
	ApiUrl       string `json:"api_url"`
	RepositoryId int64  `json:"repository_id"`
}

type CreateVolumeProvider struct {
	Type global.VolumeProviderType `json:"type"`
	Name string                    `json:"name"`

	Dboxed *CreateVolumeProviderDboxed `json:"dboxed,omitempty"`
}

type CreateVolumeProviderDboxed struct {
	ApiUrl       string `json:"api_url"`
	Token        string `json:"token"`
	RepositoryId int64  `json:"repository_id"`
}

type UpdateVolumeProvider struct {
	Dboxed *UpdateVolumeProviderDboxed `json:"dboxed,omitempty"`
}

type UpdateVolumeProviderDboxed struct {
	ApiUrl       *string `json:"api_url,omitempty"`
	Token        *string `json:"token,omitempty"`
	RepositoryId *int64  `json:"repository_id,omitempty"`
}

func VolumeProviderFromDB(v dmodel.VolumeProvider) *VolumeProvider {
	return &VolumeProvider{
		ID:        v.ID,
		Workspace: v.WorkspaceID,
		CreatedAt: v.CreatedAt,
		Type:      v.Type,
		Name:      v.Name,
		Status:    v.ReconcileStatus.ReconcileStatus,
	}
}

func VolumeProviderDboxedFromDB(v dmodel.VolumeProviderDboxed) *VolumeProviderDboxed {
	return &VolumeProviderDboxed{
		ApiUrl:       v.ApiUrl.V,
		RepositoryId: v.RepositoryId.V,
	}
}
