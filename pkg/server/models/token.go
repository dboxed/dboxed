package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type Token struct {
	ID        int64     `json:"id"`
	Workspace int64     `json:"workspace"`
	CreatedAt time.Time `json:"createdAt"`

	Name  string  `json:"name"`
	Token *string `json:"token,omitempty"`

	ForWorkspace bool   `json:"forWorkspace"`
	BoxID        *int64 `json:"boxId,omitempty"`
}

type CreateToken struct {
	Name string `json:"name"`

	ForWorkspace bool   `json:"forWorkspace,omitempty"`
	BoxID        *int64 `json:"boxId,omitempty"`
}

func TokenFromDB(v dmodel.Token, withSecret bool) Token {
	ret := Token{
		ID:           v.ID,
		Workspace:    v.WorkspaceID,
		CreatedAt:    v.CreatedAt,
		Name:         v.Name,
		ForWorkspace: v.ForWorkspace,
		BoxID:        v.BoxID,
	}
	if withSecret {
		ret.Token = &v.Token
	}
	return ret
}
