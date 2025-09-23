package dmodel

import (
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type VolumeProviderDboxed struct {
	ID querier2.NullForJoin[int64] `db:"id"`

	ApiUrl       querier2.NullForJoin[string] `db:"api_url"`
	Token        querier2.NullForJoin[string] `db:"token"`
	RepositoryId querier2.NullForJoin[int64]  `db:"repository_id"`

	Status *VolumeProviderDboxedStatus `join:"true"`
}

type VolumeProviderDboxedStatus struct {
	ID querier2.NullForJoin[int64] `db:"id"`
}

func (v *VolumeProviderDboxed) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *VolumeProviderDboxedStatus) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *VolumeProviderDboxed) Update(q *querier2.Querier, apiUrl *string, token *string, repositoryId *int64) error {
	var fields []string
	if apiUrl != nil {
		v.ApiUrl = querier2.N(*apiUrl)
		fields = append(fields, "api_url")
	}
	if token != nil {
		v.Token = querier2.N(*token)
		fields = append(fields, "token")
	}
	if repositoryId != nil {
		v.RepositoryId = querier2.N(*repositoryId)
		fields = append(fields, "repository_id")
	}
	return querier2.UpdateOneFromStruct(q, v, fields...)
}
