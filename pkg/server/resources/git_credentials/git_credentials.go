package git_credentials

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/auth_middleware"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/gobwas/glob"
)

type GitCredentialsServer struct {
}

func New() *GitCredentialsServer {
	return &GitCredentialsServer{}
}

func (s *GitCredentialsServer) Init(rootGroup huma.API, workspacesGroup huma.API) error {
	huma.Post(workspacesGroup, "/git-credentials", s.restCreateGitCredentials)
	huma.Get(workspacesGroup, "/git-credentials", s.restListGitCredentials)
	huma.Get(workspacesGroup, "/git-credentials/{id}", s.restGetGitCredentials)
	huma.Patch(workspacesGroup, "/git-credentials/{id}", s.restUpdateGitCredentials)
	huma.Delete(workspacesGroup, "/git-credentials/{id}", s.restDeleteGitCredentials)

	return nil
}

func (s *GitCredentialsServer) restCreateGitCredentials(c context.Context, i *huma_utils.JsonBody[models.CreateGitCredentials]) (*huma_utils.JsonBody[models.GitCredentials], error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	err := s.checkCredentialsValid(c, i.Body.CredentialsType, i.Body.Username, i.Body.Password, i.Body.SshKey)
	if err != nil {
		return nil, err
	}

	_, err = glob.Compile(i.Body.PathGlob, '/')
	if err != nil {
		return nil, huma.Error400BadRequest("invalid path prefix", err)
	}

	gc := &dmodel.GitCredentials{
		OwnedByWorkspace: dmodel.OwnedByWorkspace{
			WorkspaceID: w.ID,
		},
		Host:            i.Body.Host,
		PathGlob:        i.Body.PathGlob,
		CredentialsType: i.Body.CredentialsType,
		Username:        i.Body.Username,
		Password:        i.Body.Password,
		SshKey:          i.Body.SshKey,
	}

	err = gc.Create(q)
	if err != nil {
		return nil, err
	}

	m := models.GitCredentialsFromDB(*gc)
	return huma_utils.NewJsonBody(m), nil
}

func (s *GitCredentialsServer) restListGitCredentials(c context.Context, i *struct{}) (*huma_utils.List[models.GitCredentials], error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	l, err := dmodel.ListGitCredentialsForWorkspace(q, w.ID)
	if err != nil {
		return nil, err
	}

	var ret []models.GitCredentials
	for _, gc := range l {
		ret = append(ret, models.GitCredentialsFromDB(gc))
	}
	return huma_utils.NewList(ret, len(ret)), nil
}

func (s *GitCredentialsServer) restGetGitCredentials(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[models.GitCredentials], error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	gc, err := dmodel.GetGitCredentialsById(q, &w.ID, i.Id)
	if err != nil {
		return nil, err
	}

	m := models.GitCredentialsFromDB(*gc)
	return huma_utils.NewJsonBody(m), nil
}

type restUpdateGitCredentialsInput struct {
	huma_utils.IdByPath
	huma_utils.JsonBody[models.UpdateGitCredentials]
}

func (s *GitCredentialsServer) restUpdateGitCredentials(c context.Context, i *restUpdateGitCredentialsInput) (*huma_utils.JsonBody[models.GitCredentials], error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	gc, err := dmodel.GetGitCredentialsById(q, &w.ID, i.Id)
	if err != nil {
		return nil, err
	}

	err = s.checkCredentialsValid(c, gc.CredentialsType, i.Body.Username, i.Body.Password, i.Body.SshKey)
	if err != nil {
		return nil, err
	}

	switch gc.CredentialsType {
	case dmodel.GitCredentialsTypeBasicAuth:
		err = gc.UpdateBasicAuth(q, *i.Body.Username, *i.Body.Password)
		if err != nil {
			return nil, err
		}
	case dmodel.GitCredentialsTypeSshKey:
		err = gc.UpdateSshKey(q, *i.Body.SshKey)
		if err != nil {
			return nil, err
		}
	default:
		return nil, huma.Error400BadRequest("invalid credentials type")
	}

	m := models.GitCredentialsFromDB(*gc)
	return huma_utils.NewJsonBody(m), nil
}

func (s *GitCredentialsServer) restDeleteGitCredentials(c context.Context, i *huma_utils.IdByPath) (*huma_utils.Empty, error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	_, err := dmodel.GetGitCredentialsById(q, &w.ID, i.Id)
	if err != nil {
		return nil, err
	}

	err = querier.DeleteOneById[*dmodel.GitCredentials](q, i.Id)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

func (s *GitCredentialsServer) checkCredentialsValid(c context.Context, credentialsType dmodel.GitCredentialsType, username *string, password *string, sshKey *string) error {
	switch credentialsType {
	case dmodel.GitCredentialsTypeBasicAuth:
		if username == nil || password == nil || *username == "" || *password == "" {
			return huma.Error400BadRequest("username and password must be set")
		}
		if sshKey != nil {
			return huma.Error400BadRequest("ssh_key can not be set")
		}
		return nil
	case dmodel.GitCredentialsTypeSshKey:
		if username == nil || *username == "" {
			return huma.Error400BadRequest("username must be set")
		}
		if sshKey == nil || *sshKey == "" {
			return huma.Error400BadRequest("ssh_key must be set")
		}
		if password != nil {
			return huma.Error400BadRequest("password can not be set")
		}
		return nil
	default:
		return huma.Error400BadRequest("invalid credentials type")
	}
}
