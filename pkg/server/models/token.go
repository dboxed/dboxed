package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type Token struct {
	ID        string    `json:"id"`
	Workspace string    `json:"workspace"`
	CreatedAt time.Time `json:"createdAt"`

	Name  string  `json:"name"`
	Token *string `json:"token,omitempty"`

	ForWorkspace   bool    `json:"forWorkspace"`
	MachineID      *string `json:"machineId,omitempty"`
	BoxID          *string `json:"boxId,omitempty"`
	LoadBalancerId *string `json:"loadBalancerId,omitempty"`
}

type CreateToken struct {
	Name string `json:"name"`

	ForWorkspace   bool    `json:"forWorkspace,omitempty"`
	MachineID      *string `json:"machineId,omitempty"`
	BoxID          *string `json:"boxId,omitempty"`
	LoadBalancerId *string `json:"loadBalancerId,omitempty"`
}

func TokenFromDB(v dmodel.Token, withSecret bool) Token {
	ret := Token{
		ID:             v.ID,
		Workspace:      v.WorkspaceID,
		CreatedAt:      v.CreatedAt,
		Name:           v.Name,
		ForWorkspace:   v.ForWorkspace,
		MachineID:      v.MachineID,
		BoxID:          v.BoxID,
		LoadBalancerId: v.LoadBalancerId,
	}
	if withSecret {
		ret.Token = &v.Token
	}
	return ret
}
