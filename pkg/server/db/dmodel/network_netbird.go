package dmodel

import (
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type NetworkNetbird struct {
	ID querier2.NullForJoin[int64] `db:"id"`

	NetbirdVersion querier2.NullForJoin[string] `db:"netbird_version"`
	ApiUrl         querier2.NullForJoin[string] `db:"api_url"`
	ApiAccessToken *string                      `db:"api_access_token"`
}

func (v *NetworkNetbird) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *NetworkNetbird) UpdateNetbirdVersion(q *querier2.Querier, version string) error {
	v.NetbirdVersion = querier2.N(version)
	return querier2.UpdateOneFromStruct(q, v, "netbird_version")
}

func (v *NetworkNetbird) UpdateApiAccessToken(q *querier2.Querier, token *string) error {
	v.ApiAccessToken = token
	return querier2.UpdateOneFromStruct(q, v, "api_access_token")
}
