package dmodel

import "github.com/dboxed/dboxed/pkg/server/db/querier"

type Token struct {
	ID int64 `db:"id" omitCreate:"true"`
	Times

	Token string `db:"token"`

	Name   string `db:"name"`
	UserID string `db:"user_id"`

	User *User `join:"true" join_left_field:"user_id"`
}

func (v *Token) Create(q *querier.Querier) error {
	return querier.Create(q, v)
}

func GetTokenById(q *querier.Querier, userId *string, id int64) (*Token, error) {
	return querier.GetOne[Token](q, map[string]any{
		"id":      id,
		"user_id": querier.OmitIfNull(userId),
	})
}

func GetTokenByToken(q *querier.Querier, token string) (*Token, error) {
	return querier.GetOne[Token](q, map[string]any{
		"token": token,
	})
}

func ListTokensForUser(q *querier.Querier, userId string) ([]Token, error) {
	return querier.GetMany[Token](q, map[string]any{
		"user_id": userId,
	})
}
