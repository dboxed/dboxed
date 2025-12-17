package git_spec

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type UpdateCmd struct {
	GitSpec string `help:"Specify git spec ID" required:"" arg:""`

	GitUrl   *string `help:"Git repository URL"`
	Subdir   *string `help:"Subdirectory in the repository"`
	SpecFile *string `help:"Spec file name within the subdirectory"`
}

func (cmd *UpdateCmd) Run(g *flags.GlobalFlags) error {
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

	req := models.UpdateGitSpec{
		GitUrl:   cmd.GitUrl,
		Subdir:   cmd.Subdir,
		SpecFile: cmd.SpecFile,
	}

	updated, err := c2.UpdateGitSpec(ctx, gs.ID, req)
	if err != nil {
		return err
	}

	slog.Info("git spec updated", slog.Any("id", updated.ID), slog.Any("gitUrl", updated.GitUrl))

	return nil
}
