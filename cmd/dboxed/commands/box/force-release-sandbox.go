package box

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type ForceReleaseSandboxCmd struct {
	Box string `help:"Box ID or name" required:"" arg:""`
}

func (cmd *ForceReleaseSandboxCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.BoxClient{Client: c}

	b, err := commandutils.GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	if b.Sandbox != nil {
		err = c2.ReleaseSandbox(ctx, b.ID, b.Sandbox.ID)
		if err != nil {
			return err
		}
		slog.Info("sandbox detached", "box_id", b.ID, "sandbox_id", b.Sandbox.ID)
	} else {
		slog.Info("box has no sandbox ", "box_id", b.ID)
	}

	return nil
}
