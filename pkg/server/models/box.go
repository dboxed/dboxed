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
}

type UpdateBox struct {
}

type BoxToken struct {
	Token string `json:"token"`
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
