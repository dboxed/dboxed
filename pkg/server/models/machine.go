package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/global"
)

type Machine struct {
	ID        int64     `json:"id"`
	Workspace int64     `json:"workspace"`
	CreatedAt time.Time `json:"createdAt"`

	Status        string `json:"status"`
	StatusDetails string `json:"statusDetails"`

	Name string `json:"name"`

	Box int64 `json:"box"`

	MachineProvider     int64                      `json:"machineProvider"`
	MachineProviderType global.MachineProviderType `json:"machineProviderType"`
}

type CreateMachine struct {
	Name string `json:"name"`

	Box int64 `json:"box"`

	MachineProvider int64 `json:"machineProvider"`

	Hetzner *CreateMachineHetzner `json:"hetzner,omitempty"`
	Aws     *CreateMachineAws     `json:"aws,omitempty"`
}

type CreateMachineHetzner struct {
	ServerType     string `json:"serverType"`
	ServerLocation string `json:"serverLocation"`
}

type CreateMachineAws struct {
	InstanceType   string `json:"instanceType"`
	SubnetId       string `json:"subnetId"`
	RootVolumeSize *int64 `json:"rootVolumeSize,omitempty"`
}

type UpdateMachine struct {
}

func MachineFromDB(s dmodel.Machine) (*Machine, error) {
	ret := &Machine{
		ID:            s.ID,
		Workspace:     s.WorkspaceID,
		CreatedAt:     s.CreatedAt,
		Status:        s.ReconcileStatus.ReconcileStatus,
		StatusDetails: s.ReconcileStatus.ReconcileStatusDetails,

		Name: s.Name,

		Box: s.BoxID,

		MachineProvider:     s.MachineProviderID,
		MachineProviderType: global.MachineProviderType(s.MachineProviderType),
	}

	return ret, nil
}
