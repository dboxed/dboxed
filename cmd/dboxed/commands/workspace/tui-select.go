package workspace

import (
	"context"

	"github.com/charmbracelet/bubbles/list"
	"github.com/dboxed/dboxed/cmd/dboxed/cliutils"
	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type tuiWorkspaceItem struct {
	workspace *models.Workspace
}

func (i *tuiWorkspaceItem) GetId() int64 {
	return i.workspace.ID
}

func (i *tuiWorkspaceItem) GetName() string {
	return i.workspace.Name
}

func (i *tuiWorkspaceItem) FilterValue() string {
	return ""
}

func TuiSelectWorkspace(ctx context.Context, c *baseclient.Client) error {
	workspaces, err := c.ListWorkspaces(ctx)
	if err != nil {
		return err
	}

	var tuiWorkspaces []list.Item
	for _, w := range workspaces {
		tuiWorkspaces = append(tuiWorkspaces, &tuiWorkspaceItem{
			workspace: &w,
		})
	}

	selected, err := cliutils.ListSelect[*tuiWorkspaceItem]("Please select a workspace", tuiWorkspaces)
	if err != nil {
		return err
	}

	_, err = c.SwitchWorkspaceById(ctx, selected.workspace.ID)
	if err != nil {
		return err
	}

	err = c.WriteClientAuth()
	if err != nil {
		return err
	}
	return nil
}
