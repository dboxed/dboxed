package workspace

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
)

type SelectCmd struct {
	Workspace string `help:"Specify the workspace." required:"" arg:""`
}

func (cmd *SelectCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient()
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
