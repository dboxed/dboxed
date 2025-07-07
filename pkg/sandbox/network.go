package sandbox

import (
	"context"
	"fmt"
	"github.com/koobox/unboxed/pkg/util"
	net2 "github.com/koobox/unboxed/pkg/util/net"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"log/slog"
	"net"
	"os"
	"runtime"
)

const vethNamePrefix = "ub"

func (rn *Sandbox) initNetworkNames() {
	h := util.Sha256Sum([]byte(rn.SandboxName))
	h = h[:6]
	rn.NetworkNamespaceName = fmt.Sprintf("ub-%s", h)
	rn.vethNameHost = fmt.Sprintf("%s-%s-host", vethNamePrefix, h)
	rn.vethNamePeer = fmt.Sprintf("%s-%s-peer", vethNamePrefix, h)
}

func (rn *Sandbox) destroyNetworking(ctx context.Context) error {
	ns, err := netns.GetFromName(rn.NetworkNamespaceName)
	if err != nil && !os.IsNotExist(err) {
		return err
	} else if err == nil {
		slog.InfoContext(ctx, fmt.Sprintf("deleting network namespace %s", rn.NetworkNamespaceName), slog.Any("handle", ns))
		err = netns.DeleteNamed(rn.NetworkNamespaceName)
		if err != nil {
			return err
		}
	}

	l, err := netlink.LinkByName(rn.vethNameHost)
	if err == nil {
		slog.InfoContext(ctx, fmt.Sprintf("deleting network interface %s", rn.vethNameHost))
		err = netlink.LinkDel(l)
		if err != nil {
			return err
		}
	}

	l, err = netlink.LinkByName(rn.vethNamePeer)
	if err == nil {
		slog.InfoContext(ctx, fmt.Sprintf("deleting network interface %s", rn.vethNamePeer))
		err = netlink.LinkDel(l)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rn *Sandbox) setupNetworking(ctx context.Context) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hostIP, err := net2.GetIndexedIP(rn.VethNetworkCidr, 0)
	if err != nil {
		return err
	}
	peerIP, err := net2.GetIndexedIP(rn.VethNetworkCidr, 1)
	if err != nil {
		return err
	}
	hostAddr := netlink.Addr{
		IPNet: &net.IPNet{
			IP:   hostIP,
			Mask: rn.VethNetworkCidr.Mask,
		},
	}
	peerAddr := netlink.Addr{
		IPNet: &net.IPNet{
			IP:   peerIP,
			Mask: rn.VethNetworkCidr.Mask,
		},
	}

	log := slog.With(slog.Any("hostAddr", hostAddr.String()), slog.Any("peerAddr", peerAddr.String()))

	rn.HostNetworkNamespace, err = netns.Get()
	if err != nil {
		return err
	}

	log.InfoContext(ctx, fmt.Sprintf("creating network namespace %s", rn.NetworkNamespaceName))
	rn.NetworkNamespace, err = util.NewNetNsWithoutEnter(rn.NetworkNamespaceName)
	if err != nil {
		return err
	}

	log.InfoContext(ctx, fmt.Sprintf("creating veth-pair %s/%s", rn.vethNameHost, rn.vethNamePeer))
	la := netlink.NewLinkAttrs()
	la.Name = rn.vethNameHost
	veth := &netlink.Veth{
		LinkAttrs:     la,
		PeerName:      rn.vethNamePeer,
		PeerNamespace: netlink.NsFd(rn.NetworkNamespace),
	}

	err = netlink.LinkAdd(veth)
	if err != nil {
		return err
	}

	hostLink, err := netlink.LinkByName(rn.vethNameHost)
	if err != nil {
		return err
	}
	err = netlink.AddrAdd(hostLink, &hostAddr)
	if err != nil {
		return err
	}
	err = netlink.LinkSetUp(hostLink)
	if err != nil {
		return err
	}

	// route the peer veth IP into the host veth interface
	err = netlink.RouteAdd(&netlink.Route{
		Dst: &net.IPNet{
			IP:   peerAddr.IP,
			Mask: net.CIDRMask(32, 32),
		},
		LinkIndex: hostLink.Attrs().Index,
	})
	if err != nil {
		return err
	}

	var peerLink netlink.Link
	err = util.RunInNetNs(rn.NetworkNamespace, func() error {
		loLink, err := netlink.LinkByName("lo")
		if err != nil {
			return err
		}
		err = netlink.LinkSetUp(loLink)
		if err != nil {
			return err
		}

		peerLink, err = netlink.LinkByName(rn.vethNamePeer)
		if err != nil {
			return err
		}
		err = netlink.LinkSetUp(peerLink)
		if err != nil {
			return err
		}

		err = netlink.AddrAdd(peerLink, &peerAddr)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	log.InfoContext(ctx, "enabling ip forwarding")
	err = os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), 0600)
	if err != nil {
		return err
	}

	err = rn.setupIptables(ctx, hostAddr)
	if err != nil {
		return err
	}

	rn.startFixNetbirdRulesThread(ctx)
	rn.watchAndUpdatePortForwards(ctx, hostAddr, peerIP)

	err = rn.watchAndUpdateRoutes(ctx, hostAddr.IP, peerLink)
	if err != nil {
		return err
	}

	return nil
}
