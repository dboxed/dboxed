package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/util"
)

type Box struct {
	ID        string    `json:"id"`
	Workspace string    `json:"workspace"`
	CreatedAt time.Time `json:"createdAt"`

	Status        string `json:"status"`
	StatusDetails string `json:"statusDetails"`

	Name    string         `json:"name"`
	BoxType dmodel.BoxType `json:"boxType"`

	Machine *string `json:"machine"`

	Network     *string             `json:"network"`
	NetworkType *dmodel.NetworkType `json:"networkType"`

	Enabled bool `json:"enabled"`

	Sandbox *BoxSandbox `json:"sandbox,omitempty"`
}

type CreateBox struct {
	Name string `json:"name"`

	Network *string `json:"network,omitempty"`

	VolumeAttachments []AttachVolumeRequest     `json:"volumeAttachments,omitempty"`
	ComposeProjects   []CreateBoxComposeProject `json:"composeProjects,omitempty"`
}

func BoxFromDB(s dmodel.Box, sandbox *dmodel.BoxSandbox) (*Box, error) {
	var networkType *dmodel.NetworkType
	if s.NetworkType != nil {
		networkType = util.Ptr(*s.NetworkType)
	}
	ret := &Box{
		ID:            s.ID,
		Workspace:     s.WorkspaceID,
		CreatedAt:     s.CreatedAt,
		Status:        s.ReconcileStatus.ReconcileStatus.V,
		StatusDetails: s.ReconcileStatus.ReconcileStatusDetails.V,

		Name:    s.Name,
		BoxType: dmodel.BoxType(s.BoxType),

		Machine: s.MachineID,

		Network:     s.NetworkID,
		NetworkType: networkType,

		Enabled: s.Enabled,
	}

	if sandbox != nil && sandbox.ID.Valid {
		ret.Sandbox = BoxSandboxFromDB(*sandbox)
	}

	return ret, nil
}
