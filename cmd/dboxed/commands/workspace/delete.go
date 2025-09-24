package workspace

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
)

type DeleteCmd struct {
	Workspace string `help:"Specify the workspace." required:"" arg:""`
}

func (cmd *DeleteCmd) Run() error {
	ctx := context.Background()

	c, err := baseclient.FromClientAuthFile()
	if err != nil {
		return err
	}

	c2 := &clients.WorkspacesClient{Client: c}

	w, err := getWorkspace(ctx, c, cmd.Workspace)
	if err != nil {
		return err
	}

	err = c2.DeleteWorkspace(ctx, w.ID)
	if err != nil {
		return err
	}

	slog.Info("workspace deleted", slog.Any("id", w.ID))

	return nil
}
