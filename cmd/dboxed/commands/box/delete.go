package box

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type DeleteCmd struct {
	Box   string `help:"Specify the box" required:"" arg:""`
	Force bool   `help:"Skip confirmation prompt" short:"f"`
}

func (cmd *DeleteCmd) Run(g *flags.GlobalFlags) error {
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

	// Check if box is enabled and prompt for confirmation
	if b.Enabled && !cmd.Force {
		warning := fmt.Sprintf("Box '%s' (ID: %s) has desired state 'up' and may be running!\nDeleting a running box may cause data loss or disruption.", b.Name, b.ID)
		confirmed, err := commandutils.ConfirmDanger(
			"Are you sure you want to delete this box?",
			warning,
		)
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println("Deletion cancelled.")
			return nil
		}
	}

	err = c2.DeleteBox(ctx, b.ID)
	if err != nil {
		return err
	}

	slog.Info("box deleted", slog.Any("id", b.ID))

	return nil
}
