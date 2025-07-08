package network

import (
	"context"
	"fmt"
	"github.com/koobox/unboxed/pkg/util"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"log/slog"
	"net"
	"os"
	"runtime"
)

func (n *Network) Setup(ctx context.Context) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := n.initNamesAndIPs()
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "setting up networking", slog.Any("hostAddr", n.hostAddr.String()), slog.Any("peerAddr", n.peerAddr.String()))

	err = n.setupNamespaces(ctx)
	if err != nil {
		return err
	}
	hostLink, peerLink, err := n.setupVethPair(ctx)
	if err != nil {
		return err
	}

	// route the peer veth IP into the host veth interface
	err = netlink.RouteAdd(&netlink.Route{
		Dst: &net.IPNet{
			IP:   n.peerAddr.IP,
			Mask: net.CIDRMask(32, 32),
		},
		LinkIndex: hostLink.Attrs().Index,
	})
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
	}

	slog.InfoContext(ctx, "enabling ip forwarding")
	err = os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), 0600)
	if err != nil {
		return err
	}

	err = n.setupIptables(ctx)
	if err != nil {
		return err
	}

	n.startFixNetbirdRulesThread(ctx)

	err = n.watchAndUpdateRoutes(ctx, peerLink)
	if err != nil {
		return err
	}

	return nil
}

func (n *Network) setupNamespaces(ctx context.Context) error {
	var err error
	n.HostNetworkNamespace, err = netns.Get()
	if err != nil {
		return err
	}

	n.NetworkNamespace, err = netns.GetFromName(n.NetworkNamespaceName)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		slog.InfoContext(ctx, fmt.Sprintf("creating network namespace %s", n.NetworkNamespaceName))
		n.NetworkNamespace, err = util.NewNetNsWithoutEnter(n.NetworkNamespaceName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *Network) setupVethPair(ctx context.Context) (netlink.Link, netlink.Link, error) {
	hostLink, err := netlink.LinkByName(n.vethNameHost)
	if err != nil {
		if !isLinkNotFoundError(err) {
			return nil, nil, err
		}
		slog.InfoContext(ctx, fmt.Sprintf("creating veth-pair %s/%s", n.vethNameHost, n.vethNamePeer))
		la := netlink.NewLinkAttrs()
		la.Name = n.vethNameHost
		veth := &netlink.Veth{
			LinkAttrs:     la,
			PeerName:      n.vethNamePeer,
			PeerNamespace: netlink.NsFd(n.NetworkNamespace),
		}
		err = netlink.LinkAdd(veth)
		if err != nil {
			return nil, nil, err
		}
		hostLink, err = netlink.LinkByName(n.vethNameHost)
		if err != nil {
			return nil, nil, err
		}
	}

	err = setSingleAddress(ctx, hostLink, n.hostAddr)
	if err != nil {
		return nil, nil, err
	}

	err = netlink.LinkSetUp(hostLink)
	if err != nil {
		return nil, nil, err
	}

	var peerLink netlink.Link
	err = util.RunInNetNs(n.NetworkNamespace, func() error {
		loLink, err := netlink.LinkByName("lo")
		if err != nil {
			return err
		}
		err = netlink.LinkSetUp(loLink)
		if err != nil {
			return err
		}

		peerLink, err = netlink.LinkByName(n.vethNamePeer)
		if err != nil {
			return err
		}

		err = setSingleAddress(ctx, peerLink, n.peerAddr)
		if err != nil {
			return err
		}

		err = netlink.LinkSetUp(peerLink)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return hostLink, peerLink, nil
}
