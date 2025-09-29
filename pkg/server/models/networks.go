package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/global"
)

type Network struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Workspace int64     `json:"workspace"`
	Status    string    `json:"status"`

	Type global.NetworkType `json:"type"`
	Name string             `json:"name"`

	Netbird *NetworkNetbird `json:"netbird"`
}

type NetworkNetbird struct {
	NetbirdVersion string `json:"netbirdVersion"`
	ApiUrl         string `json:"apiUrl"`
}

type CreateNetwork struct {
	Type global.NetworkType `json:"type"`
	Name string             `json:"name"`

	Netbird *CreateNetworkNetbird `json:"netbird,omitempty"`
}

type CreateNetworkNetbird struct {
	NetbirdVersion string  `json:"netbirdVersion"`
	ApiUrl         *string `json:"apiUrl,omitempty"`
	ApiAccessToken string  `json:"apiAccessToken"`
}

type UpdateNetwork struct {
	Netbird *UpdateNetworkNetbird `json:"netbird,omitempty"`
}

type UpdateNetworkNetbird struct {
	NetbirdVersion *string `json:"netbirdVersion,omitempty"`
	ApiAccessToken *string `json:"apiAccessToken,omitempty"`
}

func NetworkFromDB(v dmodel.Network) *Network {
	return &Network{
		ID:        v.ID,
		CreatedAt: v.CreatedAt,
		Workspace: v.WorkspaceID,
		Type:      global.NetworkType(v.Type),
		Name:      v.Name,
		Status:    v.ReconcileStatus.ReconcileStatus,
	}
}

func NetworkNetbirdFromDB(v dmodel.NetworkNetbird) *NetworkNetbird {
	return &NetworkNetbird{
		NetbirdVersion: v.NetbirdVersion.V,
		ApiUrl:         v.ApiUrl.V,
	}
}
