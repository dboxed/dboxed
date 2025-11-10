package box

import (
	"context"
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type ListCmd struct {
	flags.ListFlags
}

type PrintBox struct {
	ID            string `col:"ID" id:"true"`
	Name          string `col:"Name"`
	Network       string `col:"Network"`
	DesiredState  string `col:"Desired State"`
	Status        string `col:"Status"`
	StatusDetails string `col:"Status Detail"`
	SandboxStatus string `col:"Sandbox Status"`
}

func (cmd *ListCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.BoxClient{Client: c}
	ct := commandutils.NewClientTool(c)

	boxes, err := c2.ListBoxes(ctx)
	if err != nil {
		return err
	}

	var table []PrintBox
	for _, b := range boxes {
		p := PrintBox{
			ID:            b.ID,
			Name:          b.Name,
			DesiredState:  b.DesiredState,
			Status:        b.Status,
			StatusDetails: b.StatusDetails,
			SandboxStatus: "-",
		}

		// Fetch run status for this box
		sandboxStatus, err := c2.GetSandboxStatus(ctx, b.ID)
		if err == nil && sandboxStatus.RunStatus != nil {
			p.SandboxStatus = *sandboxStatus.RunStatus
		}

		if b.Network != nil {
			p.Network = ct.Networks.GetColumn(ctx, *b.Network, cmd.ShowIds)
		}
		table = append(table, p)
	}

	err = commandutils.PrintTable(os.Stdout, table, cmd.ShowIds)
	if err != nil {
		return err
	}

	return nil
}
