package token

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

type PrintToken struct {
	ID      string `col:"ID" id:"true"`
	Name    string `col:"Name"`
	Type    string `col:"Type"`
	Machine string `col:"Machine"`
	Box     string `col:"Box"`
}

func (cmd *ListCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.TokenClient{Client: c}
	ct := commandutils.NewClientTool(c)

	tokens, err := c2.ListTokens(ctx)
	if err != nil {
		return err
	}

	var table []PrintToken
	for _, token := range tokens {
		p := PrintToken{
			ID:   token.ID,
			Name: token.Name,
			Type: string(token.Type),
		}
		if token.MachineID != nil {
			p.Machine = ct.Machines.GetColumn(ctx, *token.MachineID, false)
		}
		if token.BoxID != nil {
			p.Box = ct.Boxes.GetColumn(ctx, *token.BoxID, false)
		}
		table = append(table, p)
	}

	err = commandutils.PrintTable(os.Stdout, table, cmd.ShowIds)
	if err != nil {
		return err
	}

	return nil
}
