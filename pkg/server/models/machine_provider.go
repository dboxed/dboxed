package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type MachineProvider struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Workspace string    `json:"workspace"`

	Status        string `json:"status"`
	StatusDetails string `json:"statusDetails"`

	Type dmodel.MachineProviderType `json:"type"`
	Name string                     `json:"name"`

	SshKeyFingerprint *string `json:"sshKeyFingerprint"`

	Aws     *MachineProviderAws     `json:"aws,omitempty"`
	Hetzner *MachineProviderHetzner `json:"hetzner,omitempty"`
}

type CreateMachineProvider struct {
	Type dmodel.MachineProviderType `json:"type"`
	Name string                     `json:"name"`

	SshKeyPublic *string `json:"sshKeyPublic,omitempty"`

	Aws     *CreateMachineProviderAws     `json:"aws,omitempty"`
	Hetzner *CreateMachineProviderHetzner `json:"hetzner,omitempty"`
}

type CreateMachineProviderHetzner struct {
	CloudToken    string  `json:"cloudToken"`
	RobotUsername *string `json:"robotUsername,omitempty"`
	RobotPassword *string `json:"robotPassword,omitempty"`

	HetznerNetworkName string `json:"hetznerNetworkName"`
}

type UpdateMachineProvider struct {
	SshKeyPublic *string `json:"sshKeyPublic,omitempty"`

	Aws     *UpdateMachineProviderAws     `json:"aws,omitempty"`
	Hetzner *UpdateMachineProviderHetzner `json:"hetzner,omitempty"`
}

func MachineProviderFromDB(v dmodel.MachineProvider) *MachineProvider {
	return &MachineProvider{
		ID:            v.ID,
		Workspace:     v.WorkspaceID,
		CreatedAt:     v.CreatedAt,
		Status:        v.ReconcileStatus.ReconcileStatus.V,
		StatusDetails: v.ReconcileStatus.ReconcileStatusDetails.V,

		Type: v.Type,
		Name: v.Name,
	}
}
