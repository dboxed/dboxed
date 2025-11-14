//go:build linux

package network

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func Destroy(ctx context.Context, hostNetworkNamespace *netns.NsHandle, namesAndIps NamesAndIps, infraContainerRoot string) error {
	ns, err := netns.GetFromName(namesAndIps.SandboxNamespaceName)
	if err != nil && !os.IsNotExist(err) {
		return err
	} else if err == nil {
		defer ns.Close()
		slog.InfoContext(ctx, fmt.Sprintf("deleting network namespace %s", namesAndIps.SandboxNamespaceName), slog.Any("handle", ns))
		err = netns.DeleteNamed(namesAndIps.SandboxNamespaceName)
		if err != nil {
			return err
		}
	}

	l, err := netlink.LinkByName(namesAndIps.VethNameHost)
	if err == nil {
		slog.InfoContext(ctx, fmt.Sprintf("deleting network interface %s", namesAndIps.VethNameHost))
		err = netlink.LinkDel(l)
		if err != nil {
			return err
		}
	}

	l, err = netlink.LinkByName(namesAndIps.VethNamePeer)
	if err == nil {
		slog.InfoContext(ctx, fmt.Sprintf("deleting network interface %s", namesAndIps.VethNamePeer))
		err = netlink.LinkDel(l)
		if err != nil {
			return err
		}
	}

	ipt := Iptables{
		InfraContainerRoot: infraContainerRoot,
		NamesAndIps:        namesAndIps,
		Namespace:          hostNetworkNamespace,
	}
	err = ipt.runPurgeOldRules(ctx)
	if err != nil {
		return err
	}

	return nil
}
