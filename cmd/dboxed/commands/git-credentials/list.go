package git_credentials

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

type PrintGitCredentials struct {
	ID              string `col:"ID" id:"true"`
	Host            string `col:"Host"`
	PathGlob        string `col:"Path Glob"`
	CredentialsType string `col:"Type"`
}

func (cmd *ListCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.GitCredentialsClient{Client: c}

	credentials, err := c2.ListGitCredentials(ctx)
	if err != nil {
		return err
	}

	var table []PrintGitCredentials
	for _, gc := range credentials {
		table = append(table, PrintGitCredentials{
			ID:              gc.ID,
			Host:            gc.Host,
			PathGlob:        gc.PathGlob,
			CredentialsType: string(gc.CredentialsType),
		})
	}

	err = commandutils.PrintTable(os.Stdout, table, cmd.ShowIds)
	if err != nil {
		return err
	}

	return nil
}
