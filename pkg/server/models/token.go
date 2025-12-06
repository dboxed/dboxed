package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type Token struct {
	ID        string    `json:"id"`
	Workspace string    `json:"workspace"`
	CreatedAt time.Time `json:"createdAt"`

	Name       string           `json:"name"`
	Type       dmodel.TokenType `json:"type"`
	ValidUntil *time.Time       `json:"validUntil"`

	Token *string `json:"token,omitempty"`

	MachineID      *string `json:"machineId,omitempty"`
	BoxID          *string `json:"boxId,omitempty"`
	LoadBalancerId *string `json:"loadBalancerId,omitempty"`
}

type CreateToken struct {
	Name       string           `json:"name"`
	Type       dmodel.TokenType `json:"type"`
	ValidUntil *time.Time       `json:"validUntil,omitempty"`

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
		Type:           v.Type,
		ValidUntil:     v.ValidUntil,
		MachineID:      v.MachineID,
		BoxID:          v.BoxID,
		LoadBalancerId: v.LoadBalancerId,
	}
	if withSecret {
		ret.Token = &v.Token
	}
	return ret
}
