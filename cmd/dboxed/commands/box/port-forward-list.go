package box

import (
	"context"
	"fmt"
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type ListPortForwardsCmd struct {
	Box string `help:"Box ID or name" required:"" arg:""`
	flags.ListFlags
}

type PrintPortForward struct {
	ID            string `col:"ID" id:"true"`
	Description   string `col:"Description"`
	Protocol      string `col:"Protocol"`
	HostPortRange string `col:"Host Port"`
	SandboxPort   int    `col:"Sandbox Port"`
}

func (cmd *ListPortForwardsCmd) Run(g *flags.GlobalFlags) error {
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

	portForwards, err := c2.ListPortForwards(ctx, b.ID)
	if err != nil {
		return err
	}

	var table []PrintPortForward
	for _, pf := range portForwards {
		hostPortRange := fmt.Sprintf("%d", pf.HostPortFirst)
		if pf.HostPortFirst != pf.HostPortLast {
			hostPortRange = fmt.Sprintf("%d-%d", pf.HostPortFirst, pf.HostPortLast)
		}

		desc := ""
		if pf.Description != nil {
			desc = *pf.Description
		}

		table = append(table, PrintPortForward{
			ID:            pf.ID,
			Description:   desc,
			Protocol:      pf.Protocol,
			HostPortRange: hostPortRange,
			SandboxPort:   pf.SandboxPort,
		})
	}

	err = commandutils.PrintTable(os.Stdout, table, cmd.ShowIds)
	if err != nil {
		return err
	}

	return nil
}
