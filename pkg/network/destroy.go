package network

import (
	"context"
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"log/slog"
	"os"
)

func (n *Network) Destroy(ctx context.Context) error {
	var err error
	n.NamesAndIps, err = NewNamesAndIPs(n.Config)
	if err != nil {
		return err
	}

	ns, err := netns.GetFromName(n.NamesAndIps.SandboxNamespaceName)
	if err != nil && !os.IsNotExist(err) {
		return err
	} else if err == nil {
		slog.InfoContext(ctx, fmt.Sprintf("deleting network namespace %s", n.NamesAndIps.SandboxNamespaceName), slog.Any("handle", ns))
		err = netns.DeleteNamed(n.NamesAndIps.SandboxNamespaceName)
		if err != nil {
			return err
		}
	}

	l, err := netlink.LinkByName(n.NamesAndIps.VethNameHost)
	if err == nil {
		slog.InfoContext(ctx, fmt.Sprintf("deleting network interface %s", n.NamesAndIps.VethNameHost))
		err = netlink.LinkDel(l)
		if err != nil {
			return err
		}
	}

	l, err = netlink.LinkByName(n.NamesAndIps.VethNamePeer)
	if err == nil {
		slog.InfoContext(ctx, fmt.Sprintf("deleting network interface %s", n.NamesAndIps.VethNamePeer))
		err = netlink.LinkDel(l)
		if err != nil {
			return err
		}
	}

	return nil
}
