package workspace

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type WorkspaceCommands struct {
	Create CreateCmd `cmd:"" help:"Create a workspace"`
	Delete DeleteCmd `cmd:"" help:"Delete a workspace"`
	List   ListCmd   `cmd:"" help:"List workspaces"`
	Select SelectCmd `cmd:"" help:"Select a workspace"`
}

func getWorkspace(ctx context.Context, c *baseclient.Client, workspace string) (*models.Workspace, error) {
	c2 := clients.WorkspacesClient{Client: c}
	id, err := strconv.ParseInt(workspace, 10, 64)
	if err == nil {
		v, err := c2.GetWorkspaceById(ctx, id)
		if err != nil {
			return nil, err
		}
		return v, nil
	} else {
		l, err := c2.ListWorkspaces(ctx)
		if err != nil {
			return nil, err
		}
		for _, w := range l {
			if w.Name == workspace {
				return &w, nil
			}
		}
		return nil, fmt.Errorf("workspace not found")
	}
}
