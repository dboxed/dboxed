package compose

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type UpdateCmd struct {
	Box         string `help:"Box ID, UUID, or name" required:"" arg:""`
	ComposeFile string `help:"Path to docker-compose.yml file" required:"" short:"f"`
}

func (cmd *UpdateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	b, err := commandutils.GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	name, content, err := LoadComposeFileForBox(cmd.ComposeFile)
	if err != nil {
		return err
	}

	c2 := &clients.BoxClient{Client: c}

	req := models.UpdateBoxComposeProject{
		ComposeProject: string(content),
	}

	err = c2.UpdateComposeProject(ctx, b.ID, name, req)
	if err != nil {
		return err
	}

	slog.Info("compose project updated", slog.Any("box_id", b.ID), slog.Any("name", name))

	return nil
}
