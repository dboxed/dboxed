package dmodel

import (
	"github.com/dboxed/dboxed/pkg/server/db/querier"
)

type User struct {
	ID string `db:"id"`
	Times

	Name   string `db:"name"`
	Email  string `db:"email"`
	Avatar string `db:"avatar"`
}

func ListAllUsers(q *querier.Querier) ([]User, error) {
	return querier.GetMany[User](q, nil)
}

func GetUserById(q *querier.Querier, id string) (*User, error) {
	return querier.GetOne[User](q, map[string]any{
		"id": id,
	})
}

func (v *User) CreateOrUpdate(q *querier.Querier) error {
	return querier.CreateOrUpdate(q, v, "id")
}
