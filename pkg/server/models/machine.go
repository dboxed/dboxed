package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/util"
)

type Machine struct {
	ID        string    `json:"id"`
	Workspace string    `json:"workspace"`
	CreatedAt time.Time `json:"createdAt"`

	Status        string `json:"status"`
	StatusDetails string `json:"statusDetails"`

	MachineProviderStatus        string `json:"machineProviderStatus"`
	MachineProviderStatusDetails string `json:"machineProviderStatusDetails"`

	Name string `json:"name"`

	DboxedVersion string `json:"dboxedVersion"`

	MachineProvider     *string                     `json:"machineProvider,omitempty"`
	MachineProviderType *dmodel.MachineProviderType `json:"machineProviderType,omitempty"`

	Aws     *MachineAws     `json:"aws,omitempty"`
	Hetzner *MachineHetzner `json:"hetzner,omitempty"`

	RunStatus *MachineRunStatus `json:"runStatus,omitempty"`
}

type CreateMachine struct {
	Name string `json:"name"`

	MachineProvider *string `json:"machineProvider,omitempty"`

	Hetzner *CreateMachineHetzner `json:"hetzner,omitempty"`
	Aws     *CreateMachineAws     `json:"aws,omitempty"`
}

type UpdateMachine struct {
}

type AddBoxToMachineRequest struct {
	BoxId string `json:"boxId"`
}

type MachineRunStatus struct {
	StatusTime *time.Time `json:"statusTime,omitempty"`
	RunStatus  *string    `json:"runStatus,omitempty"`
	StartTime  *time.Time `json:"startTime,omitempty"`
	StopTime   *time.Time `json:"stopTime,omitempty"`
}

type UpdateMachineRunStatus struct {
	RunStatus *string    `json:"runStatus,omitempty"`
	StartTime *time.Time `json:"startTime,omitempty"`
	StopTime  *time.Time `json:"stopTime,omitempty"`
}

func MachineRunStatusFromDB(s *dmodel.MachineRunStatus) *MachineRunStatus {
	return &MachineRunStatus{
		StatusTime: s.StatusTime,
		RunStatus:  s.RunStatus,
		StartTime:  s.StartTime,
		StopTime:   s.StopTime,
	}
}

func MachineFromDB(s dmodel.Machine, runStatus *dmodel.MachineRunStatus) (*Machine, error) {
	ret := &Machine{
		ID:            s.ID,
		Workspace:     s.WorkspaceID,
		CreatedAt:     s.CreatedAt,
		Status:        s.ReconcileStatus.ReconcileStatus.V,
		StatusDetails: s.ReconcileStatus.ReconcileStatusDetails.V,

		Name:          s.Name,
		DboxedVersion: s.DboxedVersion,
	}

	if s.MachineProviderID != nil {
		ret.MachineProvider = s.MachineProviderID
		ret.MachineProviderType = util.Ptr(*s.MachineProviderType)
	}

	if s.Aws != nil && s.Aws.ID.Valid {
		ret.MachineProviderStatus = s.Aws.ReconcileStatus.ReconcileStatus.V
		ret.MachineProviderStatusDetails = s.Aws.ReconcileStatus.ReconcileStatusDetails.V
		ret.Aws = &MachineAws{
			InstanceType:   s.Aws.InstanceType.V,
			SubnetID:       s.Aws.SubnetID.V,
			RootVolumeSize: s.Aws.RootVolumeSize.V,
		}
		if s.Aws.Status != nil {
			ret.Aws.Status = &MachineStatusAws{
				InstanceId: s.Aws.Status.InstanceID,
				PublicIp4:  s.Aws.Status.PublicIp4,
			}
		}
	} else if s.Hetzner != nil && s.Hetzner.ID.Valid {
		ret.MachineProviderStatus = s.Hetzner.ReconcileStatus.ReconcileStatus.V
		ret.MachineProviderStatusDetails = s.Hetzner.ReconcileStatus.ReconcileStatusDetails.V
		ret.Hetzner = &MachineHetzner{
			ServerType:     s.Hetzner.ServerType.V,
			ServerLocation: s.Hetzner.ServerLocation.V,
		}
		if s.Hetzner.Status != nil {
			ret.Hetzner.Status = &MachineStatusHetzner{
				ServerId:  s.Hetzner.Status.ServerID,
				PublicIp4: s.Hetzner.Status.PublicIp4,
			}
		}
	}

	if runStatus != nil {
		ret.RunStatus = MachineRunStatusFromDB(runStatus)
	}

	return ret, nil
}
