package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type Token struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	Name string `json:"name"`
}

type CreateToken struct {
	Name string `json:"name"`
}

type CreateTokenResult struct {
	Token

	TokenStr string `json:"token"`
}

func TokenFromDB(v dmodel.Token) Token {
	return Token{
		ID:        v.ID,
		CreatedAt: v.CreatedAt,
		Name:      v.Name,
	}
}
