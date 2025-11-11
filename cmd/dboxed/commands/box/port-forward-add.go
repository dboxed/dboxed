package box

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type AddPortForwardCmd struct {
	Box           string  `help:"Box ID or name" required:"" arg:""`
	Description   *string `help:"Description of the port forward"`
	Protocol      string  `help:"Protocol (tcp or udp)" default:"tcp" enum:"tcp,udp"`
	HostPortFirst int     `help:"First host port" required:""`
	HostPortLast  *int    `help:"Last host port (defaults to first port)"`
	SandboxPort   int     `help:"Sandbox port" required:""`
}

func (cmd *AddPortForwardCmd) Run(g *flags.GlobalFlags) error {
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

	// Default HostPortLast to HostPortFirst if not specified
	hostPortLast := cmd.HostPortLast
	if hostPortLast == nil {
		hostPortLast = &cmd.HostPortFirst
	}

	req := models.CreateBoxPortForward{
		Description:   cmd.Description,
		Protocol:      cmd.Protocol,
		HostPortFirst: cmd.HostPortFirst,
		HostPortLast:  *hostPortLast,
		SandboxPort:   cmd.SandboxPort,
	}

	pf, err := c2.CreatePortForward(ctx, b.ID, req)
	if err != nil {
		return err
	}

	slog.Info("Created port forward", slog.Any("id", pf.ID))

	return nil
}
