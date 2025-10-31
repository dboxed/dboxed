package models

import (
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type User struct {
	ID       string  `json:"id"`
	Username string  `json:"username"`
	EMail    *string `json:"email"`
	FullName *string `json:"fullName"`
	Avatar   *string `json:"avatar"`

	IsAdmin bool `json:"isAdmin,omitempty"`
}

func UserFromDB(v dmodel.User, isAdmin bool) User {
	return User{
		ID:       v.ID,
		Username: v.Username,
		EMail:    v.EMail,
		FullName: v.FullName,
		Avatar:   v.Avatar,
		IsAdmin:  isAdmin,
	}
}
