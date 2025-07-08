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

	var err error
	n.NamesAndIps, err = NewNamesAndIPs(n.Config)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "setting up networking",
		slog.Any("hostAddr", n.NamesAndIps.HostAddr.String()),
		slog.Any("peerAddr", n.NamesAndIps.PeerAddr.String()),
	)

	err = n.setupNamespaces(ctx)
	if err != nil {
		return err
	}
	hostLink, _, err := n.setupVethPair(ctx)
	if err != nil {
		return err
	}

	// route the peer veth IP into the host veth interface
	err = netlink.RouteAdd(&netlink.Route{
		Dst: &net.IPNet{
			IP:   n.NamesAndIps.PeerAddr.IP,
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

	ipt := Iptables{
		NamesAndIps:        n.NamesAndIps,
		InfraContainerRoot: n.InfraContainerRoot,
	}

	err = ipt.setupIptables(ctx)
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

	n.NetworkNamespace, err = netns.GetFromName(n.NamesAndIps.SandboxNamespaceName)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		slog.InfoContext(ctx, fmt.Sprintf("creating network namespace %s", n.NamesAndIps.SandboxNamespaceName))
		n.NetworkNamespace, err = util.NewNetNsWithoutEnter(n.NamesAndIps.SandboxNamespaceName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *Network) setupVethPair(ctx context.Context) (netlink.Link, netlink.Link, error) {
	hostLink, err := netlink.LinkByName(n.NamesAndIps.VethNameHost)
	if err != nil {
		if !isLinkNotFoundError(err) {
			return nil, nil, err
		}
		slog.InfoContext(ctx, fmt.Sprintf("creating veth-pair %s/%s", n.NamesAndIps.VethNameHost, n.NamesAndIps.VethNamePeer))
		la := netlink.NewLinkAttrs()
		la.Name = n.NamesAndIps.VethNameHost
		veth := &netlink.Veth{
			LinkAttrs:     la,
			PeerName:      n.NamesAndIps.VethNamePeer,
			PeerNamespace: netlink.NsFd(n.NetworkNamespace),
		}
		err = netlink.LinkAdd(veth)
		if err != nil {
			return nil, nil, err
		}
		hostLink, err = netlink.LinkByName(n.NamesAndIps.VethNameHost)
		if err != nil {
			return nil, nil, err
		}
	}

	err = setSingleAddress(ctx, hostLink, n.NamesAndIps.HostAddr)
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

		peerLink, err = netlink.LinkByName(n.NamesAndIps.VethNamePeer)
		if err != nil {
			return err
		}

		err = setSingleAddress(ctx, peerLink, n.NamesAndIps.PeerAddr)
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
