package ingress_proxy

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

type PrintIngressProxy struct {
	ID            string `col:"ID" id:"true"`
	Name          string `col:"Name"`
	BoxID         string `col:"Box ID"`
	ProxyType     string `col:"Proxy Type"`
	Status        string `col:"Status"`
	StatusDetails string `col:"Status Detail"`
}

func (cmd *ListCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.IngressProxyClient{Client: c}

	proxies, err := c2.ListIngressProxies(ctx)
	if err != nil {
		return err
	}

	var table []PrintIngressProxy
	for _, p := range proxies {
		table = append(table, PrintIngressProxy{
			ID:            p.ID,
			Name:          p.Name,
			BoxID:         p.BoxID,
			ProxyType:     string(p.ProxyType),
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
