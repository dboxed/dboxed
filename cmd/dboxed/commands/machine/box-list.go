package machine

import (
	"context"
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type ListBoxesCmd struct {
	Machine string `help:"Machine ID or name" required:"" arg:""`
	flags.ListFlags
}

type PrintMachineBox struct {
	ID            string `col:"ID" id:"true"`
	Name          string `col:"Name"`
	DesiredState  string `col:"Desired State"`
	Status        string `col:"Status"`
	StatusDetails string `col:"Status Detail"`
}

func (cmd *ListBoxesCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	m, err := commandutils.GetMachine(ctx, c, cmd.Machine)
	if err != nil {
		return err
	}

	c2 := &clients.MachineClient{Client: c}

	boxes, err := c2.ListBoxes(ctx, m.ID)
	if err != nil {
		return err
	}

	var table []PrintMachineBox
	for _, b := range boxes {
		table = append(table, PrintMachineBox{
			ID:            b.ID,
			Name:          b.Name,
			DesiredState:  b.DesiredState,
			Status:        b.Status,
			StatusDetails: b.StatusDetails,
		})
	}

	err = commandutils.PrintTable(os.Stdout, table, cmd.ShowIds)
	if err != nil {
		return err
	}

	return nil
}
