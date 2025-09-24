package workspace

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/baseclient"
)

type SelectCmd struct {
	Workspace string `help:"Specify the workspace." required:"" arg:""`
}

func (cmd *SelectCmd) Run() error {
	ctx := context.Background()

	c, err := baseclient.FromClientAuthFile()
	if err != nil {
		return err
	}

	w, err := getWorkspace(ctx, c, cmd.Workspace)
	if err != nil {
		return err
	}

	_, err = c.SwitchWorkspaceById(ctx, w.ID)
	if err != nil {
		return err
	}

	slog.Info("workspace switched", slog.Any("id", w.ID))

	return nil
}
