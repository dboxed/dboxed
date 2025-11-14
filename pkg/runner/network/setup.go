//go:build linux

package network

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"runtime"

	"github.com/dboxed/dboxed/pkg/util"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func SetupSandboxNamespace(ctx context.Context, namesAndIps NamesAndIps) error {
	slog.InfoContext(ctx, "setting up sandbox netns", slog.Any("namespaceName", namesAndIps.SandboxNamespaceName))

	ns, err := netns.GetFromName(namesAndIps.SandboxNamespaceName)
	if err == nil {
		slog.InfoContext(ctx, "sandbox netns already exists")
		defer ns.Close()
	} else {
		if !os.IsNotExist(err) {
			return err
		}
		slog.InfoContext(ctx, fmt.Sprintf("creating sandbox netns %s", namesAndIps.SandboxNamespaceName))
		ns, err = util.NewNetNsWithoutEnter(namesAndIps.SandboxNamespaceName)
		if err != nil {
			return err
		}
		defer ns.Close()
	}

	err = setupSandboxLoopDevice(ctx, ns)
	if err != nil {
		return err
	}

	return nil
}

func SetupInSandbox(ctx context.Context, hostNetworkNamespace netns.NsHandle, namesAndIps NamesAndIps) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	slog.InfoContext(ctx, "setting up networking",
		slog.Any("hostAddr", namesAndIps.HostAddr.String()),
		slog.Any("peerAddr", namesAndIps.PeerAddr.String()),
	)

	err := setupVethInterfaces(ctx, hostNetworkNamespace, namesAndIps)
	if err != nil {
		return err
	}

	ipt := Iptables{
		NamesAndIps: namesAndIps,
		Namespace:   &hostNetworkNamespace,
	}

	err = ipt.setupIptables(ctx)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "enabling ip forwarding")
	err = util.RunInNetNs(hostNetworkNamespace, func() error {
		return os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), 0600)
	})
	if err != nil {
		return err
	}

	return nil
}

func setupVethInterfaces(ctx context.Context, hostNetworkNamespace netns.NsHandle, namesAndIps NamesAndIps) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	sandboxNamespace, err := netns.Get()
	if err != nil {
		return err
	}
	defer sandboxNamespace.Close()

	hostNetlink, err := netlink.NewHandleAt(hostNetworkNamespace)
	if err != nil {
		return err
	}
	sandboxNetlink, err := netlink.NewHandleAt(sandboxNamespace)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "setting up veth pair",
		slog.Any("nameHost", namesAndIps.VethNameHost),
		slog.Any("namePeer", namesAndIps.VethNamePeer),
	)

	hostLink, err := hostNetlink.LinkByName(namesAndIps.VethNameHost)
	if err == nil {
		slog.InfoContext(ctx, "veth pair already exists")
	} else {
		if !isLinkNotFoundError(err) {
			return err
		}
		slog.InfoContext(ctx, "creating veth-pair pair")
		la := netlink.NewLinkAttrs()
		la.Name = namesAndIps.VethNameHost
		la.MTU = 1384 // TODO auto detect best value
		veth := &netlink.Veth{
			LinkAttrs:     la,
			PeerName:      namesAndIps.VethNamePeer,
			PeerNamespace: netlink.NsFd(sandboxNamespace),
		}
		err = hostNetlink.LinkAdd(veth)
		if err != nil {
			return err
		}
		hostLink, err = hostNetlink.LinkByName(namesAndIps.VethNameHost)
		if err != nil {
			return err
		}
	}

	err = addAddress(ctx, hostNetlink, hostLink, namesAndIps.HostAddr, true)
	if err != nil {
		return err
	}

	if hostLink.Attrs().Flags&net.FlagUp == 0 {
		slog.InfoContext(ctx, "bringing veth host link up")
		err = hostNetlink.LinkSetUp(hostLink)
		if err != nil {
			return err
		}
	} else {
		slog.InfoContext(ctx, "veth host link is already up")
	}

	peerLink, err := sandboxNetlink.LinkByName(namesAndIps.VethNamePeer)
	if err != nil {
		return err
	}

	err = addAddress(ctx, sandboxNetlink, peerLink, namesAndIps.PeerAddr, true)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "bringing veth peer link up")
	err = sandboxNetlink.LinkSetUp(peerLink)
	if err != nil {
		return err
	}

	// route the peer veth IP into the host veth interface
	slog.InfoContext(ctx, "setting up route into namespace")
	err = hostNetlink.RouteAdd(&netlink.Route{
		Dst: &net.IPNet{
			IP:   namesAndIps.PeerAddr.IP,
			Mask: net.CIDRMask(32, 32),
		},
		LinkIndex: hostLink.Attrs().Index,
	})
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
	}

	return nil
}

func setupSandboxLoopDevice(ctx context.Context, sandboxNamespace netns.NsHandle) error {
	sandboxNetlink, err := netlink.NewHandleAt(sandboxNamespace)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "bringing lo link up")
	loLink, err := sandboxNetlink.LinkByName("lo")
	if err != nil {
		return err
	}
	if loLink.Attrs().Flags&net.FlagUp == 0 {
		err = sandboxNetlink.LinkSetUp(loLink)
		if err != nil {
			return err
		}
	}
	return nil
}

func SetupSandboxDnsProxyIp(ctx context.Context, dnsProxyIp string) error {
	sandboxNetlink, err := netlink.NewHandle()
	if err != nil {
		return err
	}

	loLink, err := sandboxNetlink.LinkByName("lo")
	if err != nil {
		return err
	}

	dnsAddr, err := netlink.ParseAddr(fmt.Sprintf("%s/8", dnsProxyIp))
	if err != nil {
		return err
	}
	err = addAddress(ctx, sandboxNetlink, loLink, *dnsAddr, false)
	if err != nil {
		return err
	}
	return nil
}
