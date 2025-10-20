package network

import (
	"context"
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type ListCmd struct{}

type PrintNetwork struct {
	ID     int64  `col:"Id"`
	Name   string `col:"Name"`
	Type   string `col:"Type"`
	Status string `col:"Status"`
}

func (cmd *ListCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.NetworkClient{Client: c}

	networks, err := c2.ListNetworks(ctx)
	if err != nil {
		return err
	}

	var table []PrintNetwork
	for _, n := range networks {
		table = append(table, PrintNetwork{
			ID:     n.ID,
			Name:   n.Name,
			Status: n.Status,
			Type:   string(n.Type),
		})
	}

	err = commandutils.PrintTable(os.Stdout, table)
	if err != nil {
		return err
	}

	return nil
}
