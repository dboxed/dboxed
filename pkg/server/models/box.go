package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/util"
)

type Box struct {
	ID        int64     `json:"id"`
	Workspace int64     `json:"workspace"`
	CreatedAt time.Time `json:"createdAt"`
	Status    string    `json:"status"`

	Uuid string `json:"uuid"`
	Name string `json:"name"`

	Machine *int64 `json:"machine"`

	Network     *int64              `json:"network"`
	NetworkType *global.NetworkType `json:"networkType"`

	DboxedVersion string `json:"dboxedVersion"`

	DesiredState string `json:"desiredState"`

	SandboxStatus *BoxSandboxStatus `json:"sandboxStatus,omitempty"`
}

type CreateBox struct {
	Name string `json:"name"`

	Network *int64 `json:"network,omitempty"`

	VolumeAttachments []AttachVolumeRequest     `json:"volumeAttachments,omitempty"`
	ComposeProjects   []CreateBoxComposeProject `json:"composeProjects,omitempty"`
}

type UpdateBox struct {
	DesiredState *string `json:"desiredState,omitempty"`
}
type BoxSandboxStatus struct {
	StatusTime *time.Time `json:"statusTime,omitempty"`
	RunStatus  *string    `json:"runStatus,omitempty"`
	StartTime  *time.Time `json:"startTime,omitempty"`
	StopTime   *time.Time `json:"stopTime,omitempty"`

	// compressed json
	DockerPs []byte `json:"dockerPs,omitempty"`
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
}

func BoxSandboxStatusFromDB(s dmodel.BoxSandboxStatus) *BoxSandboxStatus {
	return &BoxSandboxStatus{
		StatusTime: s.StatusTime,
		RunStatus:  s.RunStatus,
		StartTime:  s.StartTime,
		StopTime:   s.StopTime,
		DockerPs:   s.DockerPs,
	}
}

func BoxFromDB(s dmodel.Box, sandboxStatus *dmodel.BoxSandboxStatus) (*Box, error) {
	var networkType *global.NetworkType
	if s.NetworkType != nil {
		networkType = util.Ptr(global.NetworkType(*s.NetworkType))
	}
	ret := &Box{
		ID:        s.ID,
		Workspace: s.WorkspaceID,
		CreatedAt: s.CreatedAt,
		Status:    s.ReconcileStatus.ReconcileStatus,

		Uuid: s.Uuid,
		Name: s.Name,

		Machine: s.MachineID,

		Network:     s.NetworkID,
		NetworkType: networkType,

		DboxedVersion: s.DboxedVersion,

		DesiredState: s.DesiredState,
	}

	if sandboxStatus != nil {
		ret.SandboxStatus = BoxSandboxStatusFromDB(*sandboxStatus)
	}

	return ret, nil
}
