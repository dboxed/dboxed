package dmodel

import (
	"encoding/json"

	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/kluctl/kluctl/lib/git/types"
)

type GitSpec struct {
	OwnedByWorkspace
	SoftDeleteFields
	ReconcileStatus

	GitUrl   string  `db:"git_url"`
	GitRef   *string `db:"git_ref"`
	Subdir   string  `db:"subdir"`
	SpecFile string  `db:"spec_file"`
}

func (v *GitSpec) SetGitRef(r *types.GitRef) {
	if r == nil || *r == (types.GitRef{}) {
		v.GitRef = nil
	} else {
		b, err := json.Marshal(r)
		if err != nil {
			panic(err)
		}
		v.GitRef = util.Ptr(string(b))
	}
}

func (v *GitSpec) GetGitRef() *types.GitRef {
	if v.GitRef == nil {
		return nil
	}
	var ret types.GitRef
	err := json.Unmarshal([]byte(*v.GitRef), &ret)
	if err != nil {
		panic(err)
	}
	return &ret
}

func (v *GitSpec) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func GetGitSpecsById(q *querier2.Querier, workspaceId *string, id string, skipDeleted bool) (*GitSpec, error) {
	return querier2.GetOne[GitSpec](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           id,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	})
}

func ListGitSpecsForWorkspace(q *querier2.Querier, workspaceId string, skipDeleted bool) ([]GitSpec, error) {
	return querier2.GetMany[GitSpec](q, map[string]any{
		"workspace_id": workspaceId,
		"deleted_at":   querier2.ExcludeNonNull(skipDeleted),
	}, nil)
}

func (v *GitSpec) Update(q *querier2.Querier, gitUrl *string, gitRef **types.GitRef, subdir *string, specFile *string) error {
	var fields []string
	if gitUrl != nil {
		fields = append(fields, "git_url")
		v.GitUrl = *gitUrl
	}
	if gitRef != nil && *gitRef != nil {
		fields = append(fields, "git_ref")
		v.SetGitRef(*gitRef)
	}
	if subdir != nil {
		fields = append(fields, "subdir")
		v.Subdir = *subdir
	}
	if specFile != nil {
		fields = append(fields, "spec_file")
		v.SpecFile = *specFile
	}

	return querier2.UpdateOneByFieldsFromStruct(q, map[string]any{
		"workspace_id": v.WorkspaceID,
		"id":           v.ID,
	}, v, fields...)
}
