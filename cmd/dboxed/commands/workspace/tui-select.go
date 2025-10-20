package workspace

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/dboxed/dboxed/pkg/baseclient"
)

func TuiSelectWorkspace(ctx context.Context, c *baseclient.Client) error {
	workspaces, err := c.ListWorkspaces(ctx)
	if err != nil {
		return err
	}

	if len(workspaces) == 0 {
		return fmt.Errorf("no workspaces available")
	}

	options := make([]huh.Option[int64], len(workspaces))
	for i, w := range workspaces {
		options[i] = huh.NewOption(w.Name, w.ID)
	}

	var selectedWorkspaceID int64

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int64]().
				Title("Select a workspace").
				Options(options...).
				Value(&selectedWorkspaceID),
		),
	)

	err = form.Run()
	if err != nil {
		return err
	}

	_, err = c.SwitchWorkspaceById(ctx, selectedWorkspaceID)
	if err != nil {
		return err
	}

	err = c.WriteClientAuth()
	if err != nil {
		return err
	}
	return nil
}
