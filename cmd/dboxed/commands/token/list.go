package token

import (
	"context"
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type ListCmd struct{}

type PrintToken struct {
	ID           int64  `col:"Id"`
	Name         string `col:"Name"`
	ForWorkspace bool   `col:"For Workspace"`
	Box          string `col:"Box"`
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
			ID:           token.ID,
			Name:         token.Name,
			ForWorkspace: token.ForWorkspace,
		}
		if token.BoxID != nil {
			p.Box = ct.Boxes.GetColumn(ctx, *token.BoxID)
		}
		table = append(table)
	}

	err = commandutils.PrintTable(os.Stdout, table)
	if err != nil {
		return err
	}

	return nil
}
