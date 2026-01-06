package sandboxes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/auth_middleware"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type SandboxesServer struct {
	config config.Config
}

func New(config config.Config) *SandboxesServer {
	return &SandboxesServer{
		config: config,
	}
}

func (s *SandboxesServer) Init(rootGroup huma.API, workspacesGroup huma.API) error {
	// sandboxes
	huma.Get(workspacesGroup, "/sandboxes", s.restListSandboxes)
	huma.Get(workspacesGroup, "/sandboxes/{sandboxId}", s.restGetSandbox)

	return nil
}

func (s *SandboxesServer) restListSandboxes(c context.Context, i *huma_utils.Empty) (*huma_utils.List[models.BoxSandbox], error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	l, err := dmodel.ListSandboxesByWorkspace(q, w.ID)
	if err != nil {
		return nil, err
	}

	var ret []models.BoxSandbox
	for _, x := range l {
		ret = append(ret, *models.BoxSandboxFromDB(x))
	}

	return huma_utils.NewList(ret, len(ret)), nil
}

func (s *SandboxesServer) restGetSandbox(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[models.BoxSandbox], error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	err := auth_middleware.CheckResourceAccess(c, dmodel.TokenTypeBox, i.Id)
	if err != nil {
		return nil, err
	}

	sandbox, err := dmodel.GetSandboxById(q, &w.ID, nil, i.Id)
	if err != nil {
		return nil, err
	}

	ret := models.BoxSandboxFromDB(*sandbox)
	return huma_utils.NewJsonBody(*ret), nil
}
