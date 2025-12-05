package dmodel

import "github.com/dboxed/dboxed/pkg/server/db/querier"

type TokenType string

const (
	TokenTypeWorkspace    TokenType = "workspace"
	TokenTypeMachine      TokenType = "machine"
	TokenTypeBox          TokenType = "box"
	TokenTypeLoadBalancer TokenType = "load-balancer"
)

type Token struct {
	ID          string `db:"id" uuid:"true"`
	WorkspaceID string `db:"workspace_id"`
	Times

	Name  string    `db:"name"`
	Type  TokenType `db:"type"`
	Token string    `db:"token"`

	MachineID      *string `db:"machine_id"`
	BoxID          *string `db:"box_id"`
	LoadBalancerId *string `db:"load_balancer_id"`
}

func (v *Token) Create(q *querier.Querier) error {
	return querier.Create(q, v)
}

func GetTokenById(q *querier.Querier, workspaceId *string, id string) (*Token, error) {
	return querier.GetOne[Token](q, map[string]any{
		"workspace_id": querier.OmitIfNull(workspaceId),
		"id":           id,
	})
}

func GetTokenByName(q *querier.Querier, workspaceId string, name string) (*Token, error) {
	return querier.GetOne[Token](q, map[string]any{
		"workspace_id": workspaceId,
		"name":         name,
	})
}

func GetTokenByToken(q *querier.Querier, token string) (*Token, error) {
	return querier.GetOne[Token](q, map[string]any{
		"token": token,
	})
}

func ListTokensForWorkspace(q *querier.Querier, workspaceId string) ([]Token, error) {
	return querier.GetMany[Token](q, map[string]any{
		"workspace_id": workspaceId,
	}, nil)
}
