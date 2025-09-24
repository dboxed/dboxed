package workspace

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type CreateCmd struct {
	Name string `help:"Specify the workspace name." required:""`
}

func (cmd *CreateCmd) Run() error {
	ctx := context.Background()

	c, err := baseclient.FromClientAuthFile()
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
