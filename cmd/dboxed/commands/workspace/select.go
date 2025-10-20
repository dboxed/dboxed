package workspace

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
)

type SelectCmd struct {
	Workspace *string `help:"Specify the workspace." optional:"" arg:""`
}

func (cmd *SelectCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	if cmd.Workspace == nil {
		return TuiSelectWorkspace(ctx, c)
	}

	w, err := commandutils.GetWorkspace(ctx, c, *cmd.Workspace)
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
