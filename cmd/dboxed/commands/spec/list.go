package spec

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

type PrintDboxedSpec struct {
	ID            string `col:"ID"`
	GitUrl        string `col:"Git URL"`
	Subdir        string `col:"Subdir"`
	SpecFile      string `col:"Spec File"`
	Status        string `col:"Status"`
	StatusDetails string `col:"Status Detail"`
}

func (cmd *ListCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.DboxedSpecClient{Client: c}

	specs, err := c2.ListDboxedSpecs(ctx)
	if err != nil {
		return err
	}

	var table []PrintDboxedSpec
	for _, gs := range specs {
		table = append(table, PrintDboxedSpec{
			ID:            gs.ID,
			GitUrl:        gs.GitUrl,
			Subdir:        gs.Subdir,
			SpecFile:      gs.SpecFile,
			Status:        gs.Status,
			StatusDetails: gs.StatusDetails,
		})
	}

	err = commandutils.PrintTable(os.Stdout, table, cmd.ShowIds)
	if err != nil {
		return err
	}

	return nil
}
