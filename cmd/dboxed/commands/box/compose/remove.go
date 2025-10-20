package compose

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type RemoveCmd struct {
	Box  string `help:"Box ID, UUID, or name" required:"" arg:""`
	Name string `help:"Compose project name" required:"" arg:""`
}

func (cmd *RemoveCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	b, err := commandutils.GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	c2 := &clients.BoxClient{Client: c}

	err = c2.DeleteComposeProject(ctx, b.ID, cmd.Name)
	if err != nil {
		return err
	}

	slog.Info("compose project removed", slog.Any("box_id", b.ID), slog.Any("name", cmd.Name))

	return nil
}
