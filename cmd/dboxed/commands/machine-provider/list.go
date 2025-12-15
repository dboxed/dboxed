package machine_provider

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

type PrintMachineProvider struct {
	ID            string `col:"ID" id:"true"`
	Name          string `col:"Name"`
	Type          string `col:"Type"`
	Status        string `col:"Status"`
	StatusDetails string `col:"Status Detail"`
}

func (cmd *ListCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.MachineProviderClient{Client: c}

	providers, err := c2.ListMachineProviders(ctx)
	if err != nil {
		return err
	}

	var table []PrintMachineProvider
	for _, p := range providers {
		table = append(table, PrintMachineProvider{
			ID:            p.ID,
			Name:          p.Name,
			Type:          string(p.Type),
			Status:        p.Status,
			StatusDetails: p.StatusDetails,
		})
	}

	err = commandutils.PrintTable(os.Stdout, table, cmd.ShowIds)
	if err != nil {
		return err
	}

	return nil
}
