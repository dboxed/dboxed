package spec

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/kluctl/kluctl/lib/git/types"
)

type CreateCmd struct {
	GitUrl string `help:"Git repository URL" required:""`

	Branch *string `help:"Specify git branch" xor:"ref"`
	Tag    *string `help:"Specify git tag" xor:"ref"`
	Commit *string `help:"Specify git commit" xor:"ref"`

	Subdir   string `help:"Subdirectory in the repository"`
	SpecFile string `help:"Spec file name within the subdirectory" required:""`
}

func (cmd *CreateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.DboxedSpecClient{Client: c}

	req := models.CreateDboxedSpec{
		GitUrl:   cmd.GitUrl,
		Subdir:   cmd.Subdir,
		SpecFile: cmd.SpecFile,
	}

	if cmd.Branch != nil {
		req.GitRef = &types.GitRef{Branch: *cmd.Branch}
	} else if cmd.Tag != nil {
		req.GitRef = &types.GitRef{Tag: *cmd.Tag}
	} else if cmd.Commit != nil {
		req.GitRef = &types.GitRef{Commit: *cmd.Commit}
	}

	gs, err := c2.CreateDboxedSpec(ctx, req)
	if err != nil {
		return err
	}

	slog.Info("dboxed spec created", slog.Any("id", gs.ID), slog.Any("gitUrl", gs.GitUrl))

	return nil
}
