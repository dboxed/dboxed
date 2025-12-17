package git_spec

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type DeleteCmd struct {
	GitSpec string `help:"Specify git spec ID" required:"" arg:""`
}

func (cmd *DeleteCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	gs, err := commandutils.GetGitSpec(ctx, c, cmd.GitSpec)
	if err != nil {
		return err
	}

	c2 := &clients.GitSpecClient{Client: c}

	err = c2.DeleteGitSpec(ctx, gs.ID)
	if err != nil {
		return err
	}

	slog.Info("git spec deleted", slog.Any("id", gs.ID))

	return nil
}
