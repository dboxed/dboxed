package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/global"
)

type Machine struct {
	ID        int64     `json:"id"`
	Workspace int64     `json:"workspace"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`

	Box int64 `json:"box"`

	MachineProvider     int64                      `json:"machine_provider"`
	MachineProviderType global.MachineProviderType `json:"machine_provider_type"`
}

type CreateMachine struct {
	Name string `json:"name"`

	Box int64 `json:"box"`

	MachineProvider int64 `json:"machine_provider"`

	Hetzner *CreateMachineHetzner `json:"hetzner,omitempty"`
	Aws     *CreateMachineAws     `json:"aws,omitempty"`
}

type CreateMachineHetzner struct {
	ServerType     string `json:"server_type"`
	ServerLocation string `json:"server_location"`
}

type CreateMachineAws struct {
	InstanceType   string `json:"instance_type"`
	SubnetId       string `json:"subnet_id"`
	RootVolumeSize *int64 `json:"root_volume_size,omitempty"`
}

type UpdateMachine struct {
}

func MachineFromDB(s dmodel.Machine) (*Machine, error) {
	ret := &Machine{
		ID:        s.ID,
		Workspace: s.WorkspaceID,
		CreatedAt: s.CreatedAt,
		Name:      s.Name,

		Box: s.BoxID,

		MachineProvider:     s.MachineProviderID,
		MachineProviderType: global.MachineProviderType(s.MachineProviderType),
	}

	return ret, nil
}
