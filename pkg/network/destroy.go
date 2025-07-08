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
	ns, err := netns.GetFromName(n.NetworkNamespaceName)
	if err != nil && !os.IsNotExist(err) {
		return err
	} else if err == nil {
		slog.InfoContext(ctx, fmt.Sprintf("deleting network namespace %s", n.NetworkNamespaceName), slog.Any("handle", ns))
		err = netns.DeleteNamed(n.NetworkNamespaceName)
		if err != nil {
			return err
		}
	}

	l, err := netlink.LinkByName(n.vethNameHost)
	if err == nil {
		slog.InfoContext(ctx, fmt.Sprintf("deleting network interface %s", n.vethNameHost))
		err = netlink.LinkDel(l)
		if err != nil {
			return err
		}
	}

	l, err = netlink.LinkByName(n.vethNamePeer)
	if err == nil {
		slog.InfoContext(ctx, fmt.Sprintf("deleting network interface %s", n.vethNamePeer))
		err = netlink.LinkDel(l)
		if err != nil {
			return err
		}
	}

	return nil
}
