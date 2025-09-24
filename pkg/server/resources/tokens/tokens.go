package tokens

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/auth"
	"github.com/dboxed/dboxed/pkg/server/resources/huma_metadata"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/google/uuid"
)

type TokenServer struct {
}

func New() *TokenServer {
	s := &TokenServer{}
	return s
}

func (s *TokenServer) Init(api huma.API) error {
	huma.Post(api, "/v1/tokens", s.restCreateToken, huma_metadata.NoTokenModifier())
	huma.Get(api, "/v1/tokens", s.restListTokens, huma_metadata.NoTokenModifier())
	huma.Get(api, "/v1/tokens/{id}", s.restGetToken, huma_metadata.NoTokenModifier())
	huma.Delete(api, "/v1/tokens/{id}", s.restDeleteToken, huma_metadata.NoTokenModifier())

	return nil
}

func (s *TokenServer) restCreateToken(ctx context.Context, i *huma_utils.JsonBody[models.CreateToken]) (*huma_utils.JsonBody[models.CreateTokenResult], error) {
	q := querier.GetQuerier(ctx)
	user := auth.MustGetUser(ctx)

	err := util.CheckName(i.Body.Name)
	if err != nil {
		return nil, err
	}

	t := dmodel.Token{
		Token:  auth.TokenPrefix + uuid.NewString(),
		Name:   i.Body.Name,
		UserID: user.ID,
	}

	err = t.Create(q)
	if err != nil {
		return nil, err
	}

	return huma_utils.NewJsonBody(models.CreateTokenResult{
		Token:    models.TokenFromDB(t),
		TokenStr: t.Token,
	}), nil
}

func (s *TokenServer) restListTokens(ctx context.Context, i *struct{}) (*huma_utils.List[models.Token], error) {
	q := querier.GetQuerier(ctx)
	user := auth.MustGetUser(ctx)

	l, err := dmodel.ListTokensForUser(q, user.ID)
	if err != nil {
		return nil, err
	}

	var ret []models.Token
	for _, r := range l {
		mm := models.TokenFromDB(r)
		ret = append(ret, mm)
	}
	return huma_utils.NewList(ret, len(ret)), nil
}

func (s *TokenServer) restGetToken(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[models.Token], error) {
	q := querier.GetQuerier(c)
	user := auth.GetUser(c)

	t, err := dmodel.GetTokenById(q, &user.ID, i.Id)
	if err != nil {
		return nil, err
	}

	m := models.TokenFromDB(*t)
	return huma_utils.NewJsonBody(m), nil
}

func (s *TokenServer) restDeleteToken(c context.Context, i *huma_utils.IdByPath) (*huma_utils.Empty, error) {
	q := querier.GetQuerier(c)
	user := auth.GetUser(c)

	err := querier.DeleteOneByFields[dmodel.Token](q, map[string]any{
		"id":      i.Id,
		"user_id": user.ID,
	})
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}
