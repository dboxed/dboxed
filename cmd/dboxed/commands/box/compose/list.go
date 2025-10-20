package compose

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type ListCmd struct {
	Box string `help:"Box ID, UUID, or name" required:"" arg:""`
}

func (cmd *ListCmd) Run(g *flags.GlobalFlags) error {
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

	projects, err := c2.ListComposeProjects(ctx, b.ID)
	if err != nil {
		return err
	}

	for _, p := range projects {
		slog.Info("compose project", slog.Any("name", p.Name))
	}

	return nil
}
