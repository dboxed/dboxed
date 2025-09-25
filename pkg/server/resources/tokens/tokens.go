package tokens

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/auth"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/google/uuid"
)

type TokenServer struct {
}

func New() *TokenServer {
	s := &TokenServer{}
	return s
}

func (s *TokenServer) Init(rootGroup huma.API, workspacesGroup huma.API) error {
	huma.Post(workspacesGroup, "/tokens", s.restCreateToken)
	huma.Get(workspacesGroup, "/tokens", s.restListTokens)
	huma.Get(workspacesGroup, "/tokens/{id}", s.restGetToken)
	huma.Get(workspacesGroup, "/tokens/by-name/{tokenName}", s.restGetTokenByName)
	huma.Delete(workspacesGroup, "/tokens/{id}", s.restDeleteToken)

	return nil
}

func (s *TokenServer) restCreateToken(ctx context.Context, i *huma_utils.JsonBody[models.CreateToken]) (*huma_utils.JsonBody[models.CreateTokenResult], error) {
	q := querier.GetQuerier(ctx)
	w := global.GetWorkspace(ctx)

	err := util.CheckName(i.Body.Name)
	if err != nil {
		return nil, err
	}

	t := dmodel.Token{
		WorkspaceID: w.ID,
		Name:        i.Body.Name,
		Token:       auth.TokenPrefix + uuid.NewString(),
	}

	if i.Body.ForWorkspace {
		t.ForWorkspace = true
	} else if i.Body.BoxID != nil {
		box, err := dmodel.GetBoxById(q, &w.ID, *i.Body.BoxID, true)
		if err != nil {
			return nil, err
		}
		t.BoxID = &box.ID
	} else {
		return nil, huma.Error400BadRequest("missing details")
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
	w := global.GetWorkspace(ctx)

	l, err := dmodel.ListTokensForWorkspace(q, w.ID)
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
	w := global.GetWorkspace(c)

	t, err := dmodel.GetTokenById(q, &w.ID, i.Id)
	if err != nil {
		return nil, err
	}

	m := models.TokenFromDB(*t)
	return huma_utils.NewJsonBody(m), nil
}

type TokenName struct {
	TokenName string `path:"tokenName"`
}

func (s *TokenServer) restGetTokenByName(c context.Context, i *TokenName) (*huma_utils.JsonBody[models.Token], error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	t, err := dmodel.GetTokenByName(q, w.ID, i.TokenName)
	if err != nil {
		return nil, err
	}

	m := models.TokenFromDB(*t)
	return huma_utils.NewJsonBody(m), nil
}

func (s *TokenServer) restDeleteToken(c context.Context, i *huma_utils.IdByPath) (*huma_utils.Empty, error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	t, err := dmodel.GetTokenById(q, &w.ID, i.Id)
	if err != nil {
		return nil, err
	}
	err = querier.DeleteOneById[dmodel.Token](q, t.ID)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}
