package box

import (
	"context"
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type ListIngressesCmd struct {
	Box string `help:"Box ID or name" required:"" arg:""`
	flags.ListFlags
}

type PrintIngress struct {
	ID          string `col:"ID" id:"true"`
	Description string `col:"Description"`
	ProxyID     string `col:"Proxy ID"`
	Hostname    string `col:"Hostname"`
	PathPrefix  string `col:"Path Prefix"`
	Port        int    `col:"Port"`
}

func (cmd *ListIngressesCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	b, err := commandutils.GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	c2 := &clients.BoxClient{Client: c}

	ingresses, err := c2.ListBoxIngresses(ctx, b.ID)
	if err != nil {
		return err
	}

	var table []PrintIngress
	for _, ing := range ingresses {
		desc := ""
		if ing.Description != nil {
			desc = *ing.Description
		}

		table = append(table, PrintIngress{
			ID:          ing.ID,
			Description: desc,
			ProxyID:     ing.ProxyID,
			Hostname:    ing.Hostname,
			PathPrefix:  ing.PathPrefix,
			Port:        ing.Port,
		})
	}

	err = commandutils.PrintTable(os.Stdout, table, cmd.ShowIds)
	if err != nil {
		return err
	}

	return nil
}
