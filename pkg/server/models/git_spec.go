package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/kluctl/kluctl/lib/git/types"
)

type GitSpec struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Workspace string    `json:"workspace"`

	Status        string `json:"status"`
	StatusDetails string `json:"statusDetails"`

	GitUrl   string        `json:"gitUrl"`
	GitRef   *types.GitRef `json:"gitRef,omitempty"`
	Subdir   string        `json:"subdir"`
	SpecFile string        `json:"specFile"`
}

type CreateGitSpec struct {
	GitUrl   string        `json:"gitUrl"`
	GitRef   *types.GitRef `json:"gitRef,omitempty"`
	Subdir   string        `json:"subdir"`
	SpecFile string        `json:"specFile"`
}

type UpdateGitSpec struct {
	GitUrl   *string       `json:"gitUrl,omitempty"`
	GitRef   *types.GitRef `json:"gitRef,omitempty"`
	Subdir   *string       `json:"subdir,omitempty"`
	SpecFile *string       `json:"specFile,omitempty"`
}

func GitSpecFromDB(v dmodel.GitSpec) GitSpec {
	ret := GitSpec{
		ID:            v.ID,
		CreatedAt:     v.CreatedAt,
		Workspace:     v.WorkspaceID,
		Status:        v.ReconcileStatus.ReconcileStatus.V,
		StatusDetails: v.ReconcileStatus.ReconcileStatusDetails.V,
		GitUrl:        v.GitUrl,
		GitRef:        v.GetGitRef(),
		Subdir:        v.Subdir,
		SpecFile:      v.SpecFile,
	}
	return ret
}
