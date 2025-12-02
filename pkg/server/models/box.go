package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/util"
)

type Box struct {
	ID        string    `json:"id"`
	Workspace string    `json:"workspace"`
	CreatedAt time.Time `json:"createdAt"`

	Status        string `json:"status"`
	StatusDetails string `json:"statusDetails"`

	Name    string         `json:"name"`
	BoxType global.BoxType `json:"boxType"`

	Machine *string `json:"machine"`

	Network     *string             `json:"network"`
	NetworkType *global.NetworkType `json:"networkType"`

	DesiredState string `json:"desiredState"`

	SandboxStatus *BoxSandboxStatus `json:"sandboxStatus,omitempty"`
}

type CreateBox struct {
	Name string `json:"name"`

	Network *string `json:"network,omitempty"`

	VolumeAttachments []AttachVolumeRequest     `json:"volumeAttachments,omitempty"`
	ComposeProjects   []CreateBoxComposeProject `json:"composeProjects,omitempty"`
}

type BoxSandboxStatus struct {
	StatusTime *time.Time `json:"statusTime,omitempty"`
	RunStatus  *string    `json:"runStatus,omitempty"`
	StartTime  *time.Time `json:"startTime,omitempty"`
	StopTime   *time.Time `json:"stopTime,omitempty"`

	// compressed json
	DockerPs []byte `json:"dockerPs,omitempty"`

	NetworkIp4 *string `json:"networkIp4,omitempty"`
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

func BoxSandboxStatusFromDB(s dmodel.BoxSandboxStatus) *BoxSandboxStatus {
	return &BoxSandboxStatus{
		StatusTime: s.StatusTime,
		RunStatus:  s.RunStatus,
		StartTime:  s.StartTime,
		StopTime:   s.StopTime,
		DockerPs:   s.DockerPs,
		NetworkIp4: s.NetworkIP4,
	}
}

func BoxFromDB(s dmodel.Box, sandboxStatus *dmodel.BoxSandboxStatus) (*Box, error) {
	var networkType *global.NetworkType
	if s.NetworkType != nil {
		networkType = util.Ptr(global.NetworkType(*s.NetworkType))
	}
	ret := &Box{
		ID:            s.ID,
		Workspace:     s.WorkspaceID,
		CreatedAt:     s.CreatedAt,
		Status:        s.ReconcileStatus.ReconcileStatus.V,
		StatusDetails: s.ReconcileStatus.ReconcileStatusDetails.V,

		Name:    s.Name,
		BoxType: global.BoxType(s.BoxType),

		Machine: s.MachineID,

		Network:     s.NetworkID,
		NetworkType: networkType,

		DesiredState: s.DesiredState,
	}

	if sandboxStatus != nil {
		ret.SandboxStatus = BoxSandboxStatusFromDB(*sandboxStatus)
	}

	return ret, nil
}
