package models

import (
	"context"
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
}

type CreateBox struct {
	Name string `json:"name"`

	Network *int64 `json:"network,omitempty"`

	VolumeAttachments []AttachVolumeRequest     `json:"volumeAttachments,omitempty"`
	ComposeProjects   []CreateBoxComposeProject `json:"composeProjects,omitempty"`
}

type UpdateBox struct {
}

type BoxRunStatusInfo struct {
	RunStatus *string    `json:"runStatus,omitempty"`
	StartTime *time.Time `json:"startTime,omitempty"`
	StopTime  *time.Time `json:"stopTime,omitempty"`
}

type UpdateBoxRunStatus struct {
	RunStatus *BoxRunStatusInfo `json:"runStatus,omitempty"`

	// compressed json
	DockerPs []byte `json:"dockerPs,omitempty"`
}

type BoxRunStatus struct {
	StatusTime *time.Time `json:"statusTime,omitempty"`
	RunStatus  *string    `json:"runStatus,omitempty"`
	StartTime  *time.Time `json:"startTime,omitempty"`
	StopTime   *time.Time `json:"stopTime,omitempty"`

	// compressed json
	DockerPs []byte `json:"dockerPs,omitempty"`
}

func BoxRunStatusFromDB(s dmodel.BoxRunStatus) *BoxRunStatus {
	return &BoxRunStatus{
		StatusTime: s.StatusTime,
		RunStatus:  s.RunStatus,
		StartTime:  s.StartTime,
		StopTime:   s.StopTime,
		DockerPs:   s.DockerPs,
	}
}

func BoxFromDB(ctx context.Context, s dmodel.Box) (*Box, error) {
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
	}

	return ret, nil
}
