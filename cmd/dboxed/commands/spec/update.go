package spec

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type UpdateCmd struct {
	DboxedSpec string `help:"Specify dboxed spec" required:"" arg:""`

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

	gs, err := commandutils.GetDboxedSpec(ctx, c, cmd.DboxedSpec)
	if err != nil {
		return err
	}

	c2 := &clients.DboxedSpecClient{Client: c}

	req := models.UpdateDboxedSpec{
		GitUrl:   cmd.GitUrl,
		Subdir:   cmd.Subdir,
		SpecFile: cmd.SpecFile,
	}

	updated, err := c2.UpdateDboxedSpec(ctx, gs.ID, req)
	if err != nil {
		return err
	}

	slog.Info("dboxed spec updated", slog.Any("id", updated.ID), slog.Any("gitUrl", updated.GitUrl))

	return nil
}
