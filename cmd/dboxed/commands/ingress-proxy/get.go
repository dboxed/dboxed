package ingress_proxy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type GetCmd struct {
	Id string `help:"Ingress proxy ID" required:"" arg:""`
}

func (cmd *GetCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	c2 := &clients.IngressProxyClient{Client: c}

	proxy, err := c2.GetIngressProxyById(ctx, cmd.Id)
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(proxy, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(b))

	return nil
}
