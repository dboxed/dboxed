package dmodel

import querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"

type GitSpecMapping struct {
	OwnedByWorkspace

	RepoKey     string `db:"repo_key"`
	RecreateKey string `db:"recreate_key"`
	ObjectType  string `db:"object_type"`
	ObjectId    string `db:"object_id"`
	ObjectName  string `db:"object_name"`

	Spec string `db:"spec"`
}

func (v *GitSpecMapping) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func GetGitSpecMappingById(q *querier2.Querier, workspaceId *string, id string) (*GitSpecMapping, error) {
	return querier2.GetOne[GitSpecMapping](q, map[string]any{
		"workspace_id": querier2.OmitIfNull(workspaceId),
		"id":           id,
	})
}

func ListGitSpecMappingForRepoKey(q *querier2.Querier, workspaceId string, repoKey string) ([]GitSpecMapping, error) {
	return querier2.GetMany[GitSpecMapping](q, map[string]any{
		"workspace_id": workspaceId,
		"repo_key":     repoKey,
	}, nil)
}
