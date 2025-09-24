package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type Workspace struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Name      string    `json:"name"`

	Access []WorkspaceAccess `json:"access"`
}

type CreateWorkspace struct {
	Name string `json:"name"`
}

type WorkspaceIdByPath struct {
	WorkspaceId int64 `path:"workspaceId"`
}

type WorkspaceAccess struct {
	User User `json:"user"`
}

func WorkspaceFromDB(v dmodel.Workspace) Workspace {
	var access []WorkspaceAccess
	for _, a := range v.Access {
		access = append(access, WorkspaceAccess{
			User: UserFromDB(a.User, false),
		})
	}
	return Workspace{
		ID:        v.ID,
		CreatedAt: v.CreatedAt,
		Name:      v.Name,
		Access:    access,
	}
}
