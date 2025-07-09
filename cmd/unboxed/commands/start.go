package commands

import (
	"context"
	"github.com/koobox/unboxed/cmd/unboxed/flags"
	"github.com/koobox/unboxed/pkg/start-box"
	"net"
)

type StartCmd struct {
	flags.BoxUrlFlags

	BoxName     string `help:"Specify the box name" required:""`
	VethCidrArg string `help:"CIDR to use for veth pair" default:"1.2.3.0/30"`
}

func (cmd *StartCmd) Run(g *flags.GlobalFlags) error {
	url, err := cmd.GetBoxUrl()
	if err != nil {
		return err
	}

	_, vethCidr, err := net.ParseCIDR(cmd.VethCidrArg)
	if err != nil {
		return err
	}

	startBox := start_box.StartBox{
		BoxUrl:          url,
		BoxName:         cmd.BoxName,
		WorkDir:         g.WorkDir,
		VethNetworkCidr: vethCidr,
	}

	ctx := context.Background()
	err = startBox.Start(ctx)
	if err != nil {
		return err
	}

	return nil
}
