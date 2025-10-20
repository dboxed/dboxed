package compose

import (
	"context"
	"log/slog"
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type CreateCmd struct {
	Box  string `help:"Specify the box" required:"" arg:""`
	Name string `help:"Compose project name" required:"" arg:""`
	File string `help:"Path to docker-compose.yml file" required:"" short:"f"`
}

func (cmd *CreateCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	b, err := commandutils.GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(cmd.File)
	if err != nil {
		return err
	}

	c2 := &clients.BoxClient{Client: c}

	req := models.CreateBoxComposeProject{
		Name:           cmd.Name,
		ComposeProject: string(content),
	}

	err = c2.CreateComposeProject(ctx, b.ID, req)
	if err != nil {
		return err
	}

	slog.Info("compose project created", slog.Any("box_id", b.ID), slog.Any("name", cmd.Name))

	return nil
}
