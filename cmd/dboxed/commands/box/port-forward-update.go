package box

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type UpdatePortForwardCmd struct {
	Box           string  `help:"Box ID or name" required:"" arg:""`
	PortForwardId string  `help:"Port forward ID" required:"" arg:""`
	Description   *string `help:"Description of the port forward"`
	Protocol      *string `help:"Protocol (tcp or udp)" enum:"tcp,udp"`
	HostPortFirst *int    `help:"First host port"`
	HostPortLast  *int    `help:"Last host port"`
	SandboxPort   *int    `help:"Sandbox port"`
}

func (cmd *UpdatePortForwardCmd) Run(g *flags.GlobalFlags) error {
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

	req := models.UpdateBoxPortForward{
		Description:   cmd.Description,
		Protocol:      cmd.Protocol,
		HostPortFirst: cmd.HostPortFirst,
		HostPortLast:  cmd.HostPortLast,
		SandboxPort:   cmd.SandboxPort,
	}

	pf, err := c2.UpdatePortForward(ctx, b.ID, cmd.PortForwardId, req)
	if err != nil {
		return err
	}

	slog.Info("Updated port forward", slog.Any("id", pf.ID))

	return nil
}
