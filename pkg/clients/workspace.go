package clients

import (
	"context"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type WorkspacesClient struct {
	Client *baseclient.Client
}

func (c *WorkspacesClient) CreateWorkspace(ctx context.Context, req models.CreateWorkspace) (*models.Workspace, error) {
	p, err := c.Client.BuildApiPath(false, "workspaces")
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Workspace](ctx, c.Client, "POST", p, req)
}

func (c *WorkspacesClient) DeleteWorkspace(ctx context.Context, workspaceId int64) error {
	p, err := c.Client.BuildApiPath(false, "workspaces", workspaceId)
	if err != nil {
		return err
	}
	_, err = baseclient.RequestApi[huma_utils.Empty](ctx, c.Client, "DELETE", p, struct{}{})
	return err
}

func (c *WorkspacesClient) ListWorkspaces(ctx context.Context) ([]models.Workspace, error) {
	p, err := c.Client.BuildApiPath(false, "workspaces")
	if err != nil {
		return nil, err
	}
	l, err := baseclient.RequestApi[huma_utils.ListBody[models.Workspace]](ctx, c.Client, "GET", p, struct{}{})
	if err != nil {
		return nil, err
	}
	return l.Items, err
}

func (c *WorkspacesClient) GetWorkspaceById(ctx context.Context, workspaceId int64) (*models.Workspace, error) {
	p, err := c.Client.BuildApiPath(false, "workspaces", workspaceId)
	if err != nil {
		return nil, err
	}
	return baseclient.RequestApi[models.Workspace](ctx, c.Client, "GET", p, struct{}{})
}
