package network

import (
	"context"
	"fmt"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"log/slog"
	"net"
	"os"
	"runtime"
)

func (n *Network) InitNamesAndIPs() error {
	var err error
	n.NamesAndIps, err = NewNamesAndIPs(n.Config)
	if err != nil {
		return err
	}
	return nil
}

func (n *Network) Setup(ctx context.Context) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	slog.InfoContext(ctx, "setting up networking",
		slog.Any("hostAddr", n.NamesAndIps.HostAddr.String()),
		slog.Any("peerAddr", n.NamesAndIps.PeerAddr.String()),
	)

	hostLink, _, err := n.setupVethPair(ctx)
	if err != nil {
		return err
	}

	// route the peer veth IP into the host veth interface
	slog.InfoContext(ctx, "setting up route into namespace")
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

	ipt := Iptables{
		NamesAndIps: n.NamesAndIps,
	}

	err = ipt.setupIptables(ctx)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "enabling ip forwarding")
	err = os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), 0600)
	if err != nil {
		return err
	}

	return nil
}

func (n *Network) SetupNamespaces(ctx context.Context) error {
	slog.InfoContext(ctx, "setting up network namespace", slog.Any("namespaceName", n.NamesAndIps.SandboxNamespaceName))

	var err error
	n.HostNetworkNamespace, err = netns.Get()
	if err != nil {
		return err
	}

	n.NetworkNamespace, err = netns.GetFromName(n.NamesAndIps.SandboxNamespaceName)
	if err == nil {
		slog.InfoContext(ctx, "network namespace already exists")
	} else {
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
	slog.InfoContext(ctx, "setting up veth pair",
		slog.Any("nameHost", n.NamesAndIps.VethNameHost),
		slog.Any("namePeer", n.NamesAndIps.VethNamePeer),
	)

	hostLink, err := netlink.LinkByName(n.NamesAndIps.VethNameHost)
	if err == nil {
		slog.InfoContext(ctx, "veth pair already exists")
	} else {
		if !isLinkNotFoundError(err) {
			return nil, nil, err
		}
		slog.InfoContext(ctx, "creating veth-pair pair")
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

	if hostLink.Attrs().Flags&net.FlagUp == 0 {
		slog.InfoContext(ctx, "bringing veth host link up")
		err = netlink.LinkSetUp(hostLink)
		if err != nil {
			return nil, nil, err
		}
	} else {
		slog.InfoContext(ctx, "veth host link is already up")
	}

	var peerLink netlink.Link
	err = util.RunInNetNs(n.NetworkNamespace, func() error {
		slog.InfoContext(ctx, "bringing lo link up")

		loLink, err := netlink.LinkByName("lo")
		if err != nil {
			return err
		}
		if loLink.Attrs().Flags&net.FlagUp == 0 {
			err = netlink.LinkSetUp(loLink)
			if err != nil {
				return err
			}
		}

		peerLink, err = netlink.LinkByName(n.NamesAndIps.VethNamePeer)
		if err != nil {
			return err
		}

		err = setSingleAddress(ctx, peerLink, n.NamesAndIps.PeerAddr)
		if err != nil {
			return err
		}

		slog.InfoContext(ctx, "bringing veth peer link up")
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
