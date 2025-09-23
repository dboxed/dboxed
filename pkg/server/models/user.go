package models

import (
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type User struct {
	ID     string `json:"id"`
	EMail  string `json:"email"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`

	IsAdmin bool `json:"isAdmin,omitempty"`
}

func UserFromDB(v dmodel.User, isAdmin bool) User {
	return User{
		ID:      v.ID,
		EMail:   v.Email,
		Name:    v.Name,
		Avatar:  v.Avatar,
		IsAdmin: isAdmin,
	}
}
