package users

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/auth"
	"github.com/dboxed/dboxed/pkg/server/resources/huma_metadata"
)

type Users struct {
	api huma.API
}

func New() *Users {
	return &Users{}
}

func (s *Users) Init(api huma.API) error {
	s.api = api

	skipWorkspaceModifier := huma_utils.MetadataModifier(huma_metadata.SkipWorkspace, true)

	huma.Get(s.api, "/v1/admin/users", s.restListUsers, skipWorkspaceModifier, huma_metadata.NeedAdminModifier())
	huma.Get(s.api, "/v1/admin/users/{id}", s.restGetUser, skipWorkspaceModifier, huma_metadata.NeedAdminModifier())

	return nil
}

func (s *Users) restListUsers(ctx context.Context, i *struct{}) (*huma_utils.List[models.User], error) {
	q := querier.GetQuerier(ctx)

	l, err := dmodel.ListAllUsers(q)
	if err != nil {
		return nil, err
	}

	var ret []models.User
	for _, u := range l {
		um := models.UserFromDB(u)
		um.IsAdmin = auth.IsAdminUser(ctx, &um)
		ret = append(ret, um)
	}
	return huma_utils.NewList(ret, len(ret)), nil
}

func (s *Users) restGetUser(ctx context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[models.User], error) {
	q := querier.GetQuerier(ctx)

	v, err := dmodel.GetUserById(q, i.Id)
	if err != nil {
		return nil, err
	}
	um := models.UserFromDB(*v)
	um.IsAdmin = auth.IsAdminUser(ctx, &um)
	return huma_utils.NewJsonBody(um), nil
}
