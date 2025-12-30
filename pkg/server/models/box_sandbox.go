package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type BoxSandbox struct {
	ID string `json:"id"`

	MachineID string `json:"machineId"`
	Hostname  string `json:"hostname"`

	StatusTime *time.Time `json:"statusTime,omitempty"`
	RunStatus  *string    `json:"runStatus,omitempty"`
	StartTime  *time.Time `json:"startTime,omitempty"`
	StopTime   *time.Time `json:"stopTime,omitempty"`

	// compressed json
	DockerPs []byte `json:"dockerPs,omitempty"`

	NetworkIp4 *string `json:"networkIp4,omitempty"`
}

type CreateBoxSandbox struct {
	// This is the ID found in /etc/machine-id, not the dboxed machine id
	MachineId string `json:"machineId"`
	Hostname  string `json:"hostname"`
}

type UpdateBoxSandboxStatus struct {
	SandboxStatus *UpdateBoxSandboxStatus2 `json:"sandboxStatus,omitempty"`

	// compressed json
	DockerPs []byte `json:"dockerPs,omitempty"`
}

type UpdateBoxSandboxStatus2 struct {
	RunStatus *string    `json:"runStatus,omitempty"`
	StartTime *time.Time `json:"startTime,omitempty"`
	StopTime  *time.Time `json:"stopTime,omitempty"`

	NetworkIp4 *string `json:"networkIp4,omitempty"`
}

func BoxSandboxFromDB(s dmodel.BoxSandbox) *BoxSandbox {
	return &BoxSandbox{
		ID:         s.ID.V,
		MachineID:  s.MachineId.V,
		Hostname:   s.Hostname.V,
		StatusTime: s.StatusTime,
		RunStatus:  s.RunStatus,
		StartTime:  s.StartTime,
		StopTime:   s.StopTime,
		DockerPs:   s.DockerPs,
		NetworkIp4: s.NetworkIP4,
	}
}
