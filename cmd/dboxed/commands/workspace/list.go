package workspace

import (
	"context"
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type ListCmd struct {
}

type PrintWorkspace struct {
	ID            string `col:"ID"`
	Name          string `col:"Name"`
	Status        string `col:"Status"`
	StatusDetails string `col:"Status Detail"`
}

func (cmd *ListCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.WorkspacesClient{Client: c}

	l, err := c2.ListWorkspaces(ctx)
	if err != nil {
		return err
	}

	// Get the currently selected workspace ID
	currentWorkspaceId := c.GetWorkspaceId()

	var table []PrintWorkspace
	for _, w := range l {
		name := w.Name
		if currentWorkspaceId != nil && *currentWorkspaceId == w.ID {
			name += " (current)"
		}
		table = append(table, PrintWorkspace{
			ID:            w.ID,
			Name:          name,
			Status:        w.Status,
			StatusDetails: w.StatusDetails,
		})
	}

	err = commandutils.PrintTable(os.Stdout, table)
	if err != nil {
		return err
	}

	return nil
}
