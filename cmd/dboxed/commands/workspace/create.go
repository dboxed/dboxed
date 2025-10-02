package workspace

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type CreateCmd struct {
	Name string `help:"Specify the workspace name." required:""`
}

func (cmd *CreateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.WorkspacesClient{Client: c}

	rep, err := c2.CreateWorkspace(ctx, models.CreateWorkspace{
		Name: cmd.Name,
	})

	slog.Info("workspace created", slog.Any("id", rep.ID))

	return nil
}
