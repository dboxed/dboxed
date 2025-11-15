package tokens

import (
	"context"
	"fmt"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/auth"
	"github.com/dboxed/dboxed/pkg/util"
)

const InternalTokenNamePrefix = "internal_"

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

func (s *TokenServer) restCreateToken(ctx context.Context, i *huma_utils.JsonBody[models.CreateToken]) (*huma_utils.JsonBody[models.Token], error) {
	w := global.GetWorkspace(ctx)
	ret, err := CreateToken(ctx, w.ID, i.Body, true, false)
	if err != nil {
		return nil, err
	}

	return huma_utils.NewJsonBody(*ret), nil
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
		mm := models.TokenFromDB(r, false)
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

	m := models.TokenFromDB(*t, false)
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

	m := models.TokenFromDB(*t, false)
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

func CreateToken(ctx context.Context, workspaceId string, ct models.CreateToken, returnSecret bool, internal bool) (*models.Token, error) {
	q := querier.GetQuerier(ctx)

	err := util.CheckName(ct.Name)
	if err != nil {
		return nil, err
	}

	_, err = dmodel.GetTokenByName(q, workspaceId, ct.Name)
	if err != nil {
		if !querier.IsSqlNotFoundError(err) {
			return nil, err
		}
	} else {
		return nil, huma.Error409Conflict(fmt.Sprintf("token with name '%s' already exists", ct.Name))
	}

	hasInternalPrefix := strings.HasPrefix(ct.Name, InternalTokenNamePrefix)
	if !internal && hasInternalPrefix {
		return nil, huma.Error400BadRequest("invalid token name")
	} else if internal && !hasInternalPrefix {
		return nil, huma.Error400BadRequest("missing internal token prefix")
	}

	t := dmodel.Token{
		WorkspaceID: workspaceId,
		Name:        ct.Name,
		Token:       auth.TokenPrefix + util.RandomString(16),
	}

	if ct.ForWorkspace {
		t.ForWorkspace = true
	} else if ct.BoxID != nil {
		box, err := dmodel.GetBoxById(q, &workspaceId, *ct.BoxID, true)
		if err != nil {
			return nil, err
		}
		t.BoxID = &box.ID
	} else if ct.LoadBalancerId != nil {
		lb, err := dmodel.GetLoadBalancerById(q, &workspaceId, *ct.LoadBalancerId, true)
		if err != nil {
			return nil, err
		}
		t.LoadBalancerId = &lb.ID
	} else {
		return nil, huma.Error400BadRequest("missing details")
	}

	err = t.Create(q)
	if err != nil {
		return nil, err
	}

	ret := models.TokenFromDB(t, returnSecret)
	return &ret, nil
}
