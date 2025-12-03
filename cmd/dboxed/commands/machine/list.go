package machine

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

type PrintMachine struct {
	ID            string `col:"ID" id:"true"`
	Name          string `col:"Name"`
	Status        string `col:"Status"`
	StatusDetails string `col:"Status Detail"`
}

func (cmd *ListCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.MachineClient{Client: c}

	machines, err := c2.ListMachines(ctx)
	if err != nil {
		return err
	}

	var table []PrintMachine
	for _, m := range machines {
		table = append(table, PrintMachine{
			ID:            m.ID,
			Name:          m.Name,
			Status:        m.Status,
			StatusDetails: m.StatusDetails,
		})
	}

	err = commandutils.PrintTable(os.Stdout, table, cmd.ShowIds)
	if err != nil {
		return err
	}

	return nil
}
