package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type GitCredentials struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Workspace string    `json:"workspace"`

	Host            string                    `json:"host"`
	PathGlob        string                    `json:"pathGlob"`
	CredentialsType dmodel.GitCredentialsType `json:"credentialsType"`
}

type CreateGitCredentials struct {
	Host     string `json:"host"`
	PathGlob string `json:"pathGlob"`

	CredentialsType dmodel.GitCredentialsType `json:"credentialsType"`

	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
	SshKey   *string `json:"sshKey,omitempty"`
}

type UpdateGitCredentials struct {
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
	SshKey   *string `json:"sshKey,omitempty"`
}

func GitCredentialsFromDB(v dmodel.GitCredentials) GitCredentials {
	return GitCredentials{
		ID:              v.ID,
		CreatedAt:       v.CreatedAt,
		Workspace:       v.WorkspaceID,
		Host:            v.Host,
		PathGlob:        v.PathGlob,
		CredentialsType: v.CredentialsType,
	}
}
