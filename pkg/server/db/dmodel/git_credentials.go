package dmodel

import (
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type GitCredentialsType string

const (
	GitCredentialsTypeBasicAuth GitCredentialsType = "basic-auth"
	GitCredentialsTypeSshKey    GitCredentialsType = "ssh-key"
)

type GitCredentials struct {
	OwnedByWorkspace

	Host     string `db:"host"`
	PathGlob string `db:"path_glob"`

	CredentialsType GitCredentialsType `db:"credentials_type"`
	Username        *string            `db:"username"`
	Password        *string            `db:"password"`
	SshKey          *string            `db:"ssh_key"`
}

func (v *GitCredentials) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func GetGitCredentialsById(q *querier2.Querier, workspaceId *string, id string) (*GitCredentials, error) {
	return querier2.GetOne[GitCredentials](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           id,
	})
}

func ListGitCredentialsForWorkspace(q *querier2.Querier, workspaceId string) ([]GitCredentials, error) {
	return querier2.GetMany[GitCredentials](q, map[string]any{
		"workspace_id": workspaceId,
	}, nil)
}

func (v *GitCredentials) UpdateBasicAuth(q *querier2.Querier, username string, password string) error {
	v.Username = &username
	v.Password = &password
	return querier2.UpdateOneByFieldsFromStruct(q, map[string]any{
		"workspace_id": v.WorkspaceID,
		"id":           v.ID,
	}, v, "username", "password")
}

func (v *GitCredentials) UpdateSshKey(q *querier2.Querier, sshKey string) error {
	v.SshKey = &sshKey
	return querier2.UpdateOneByFieldsFromStruct(q, map[string]any{
		"workspace_id": v.WorkspaceID,
		"id":           v.ID,
	}, v, "ssh_key")
}
