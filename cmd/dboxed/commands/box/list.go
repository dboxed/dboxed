package box

import (
	"context"
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type ListCmd struct{}

type PrintBox struct {
	ID           int64  `col:"ID"`
	Name         string `col:"Name"`
	Network      string `col:"Network"`
	DesiredState string `col:"Desired State"`
	RunStatus    string `col:"Run Status"`
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
			ID:           b.ID,
			Name:         b.Name,
			DesiredState: b.DesiredState,
			RunStatus:    "-",
		}

		// Fetch run status for this box
		runStatus, err := c2.GetBoxRunStatus(ctx, b.ID)
		if err == nil && runStatus.RunStatus != nil {
			p.RunStatus = *runStatus.RunStatus
		}

		if b.Network != nil {
			p.Network = ct.Networks.GetColumn(ctx, *b.Network)
		}
		table = append(table, p)
	}

	err = commandutils.PrintTable(os.Stdout, table)
	if err != nil {
		return err
	}

	return nil
}
