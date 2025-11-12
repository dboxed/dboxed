//go:build linux

package network

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"
)

type RoutesMirror struct {
	NetworkConfig *boxspec.NetworkConfig
	HostNamespace netns.NsHandle

	namesAndIps    NamesAndIps
	hostNetlink    *netlink.Handle
	sandboxNetlink *netlink.Handle
}

func (n *RoutesMirror) Start(ctx context.Context) error {
	slog.InfoContext(ctx, "starting routes mirror")

	var err error
	n.namesAndIps, err = NewNamesAndIPs(n.NetworkConfig.SandboxName, n.NetworkConfig.VethNetworkCidr)
	if err != nil {
		return err
	}

	n.hostNetlink, err = netlink.NewHandleAt(n.HostNamespace)
	if err != nil {
		return err
	}
	n.sandboxNetlink, err = netlink.NewHandle()
	if err != nil {
		return err
	}

	peerLink, err := n.sandboxNetlink.LinkByName(n.namesAndIps.VethNamePeer)
	if err != nil {
		return err
	}

	err = n.startWatchAndUpdateRoutes(ctx, peerLink)
	if err != nil {
		return err
	}

	return nil
}

// startWatchAndUpdateRoutes will watch for routes on the host network namespace and create mirrored routes inside the dboxed
// network namespace. Each such route uses the host veth interface as gateway (NAT). A simpler solution would be to just
// add a single default route, but this would not respect differences in MTUs per host network interface.
func (n *RoutesMirror) startWatchAndUpdateRoutes(ctx context.Context, peerLink netlink.Link) error {
	routeUpdateChan := make(chan netlink.RouteUpdate)
	err := netlink.RouteSubscribeWithOptions(routeUpdateChan, ctx.Done(), netlink.RouteSubscribeOptions{
		Namespace:    &n.HostNamespace,
		ListExisting: true,
		ErrorCallback: func(err error) {
			slog.ErrorContext(ctx, "error in RouteSubscribeWithOptions", slog.Any("error", err))
		},
	})
	if err != nil {
		return err
	}

	doWork := func(tc <-chan time.Time) {
		for {
			select {
			case ru := <-routeUpdateChan:
				err := n.updateRoute(ctx, ru, peerLink)
				if err != nil && !os.IsExist(err) {
					slog.ErrorContext(ctx, "failed to update route", slog.Any("error", err))
				}
			case <-ctx.Done():
				return
			case <-tc:
				return
			}
		}
	}

	// first, proceed until no route updates happen for at least a second. This ensures that basic routes are ready.
	doWork(time.After(time.Second))

	// now proceed with the rest until the context is cancelled
	go func() {
		doWork(nil)
	}()

	return nil
}

func (n *RoutesMirror) updateRoute(ctx context.Context, ru netlink.RouteUpdate, peerLink netlink.Link) error {
	isInternalIp := func(ip net.IP) bool {
		if n.namesAndIps.HostAddr.IP.Equal(ip) {
			return true
		}
		if n.namesAndIps.PeerAddr.IP.Equal(ip) {
			return true
		}
		return false
	}

	if ru.Dst != nil {
		if ru.Dst.IP.IsLoopback() {
			return nil
		}
		if isInternalIp(ru.Dst.IP) {
			return nil
		}
		if ru.Dst.IP.To4() == nil {
			return nil
		}
	}
	if ru.Src != nil {
		if isInternalIp(ru.Src) {
			return nil
		}
	}

	hostIP := n.namesAndIps.HostAddr.IP

	hostLinks, err := n.hostNetlink.LinkList()
	if err != nil {
		return err
	}

	var hostVethLink netlink.Link
	for _, l := range hostLinks {
		if l.Attrs().Name == n.namesAndIps.VethNameHost {
			hostVethLink = l
		}
	}
	if hostVethLink == nil {
		return fmt.Errorf("link %s not found in host netns", n.namesAndIps.VethNameHost)
	}

	findLink := func(idx int) netlink.Link {
		for _, l := range hostLinks {
			if l.Attrs().Index == idx {
				return l
			}
		}
		return nil
	}
	getMtu := func(l netlink.Link) int {
		mtu := l.Attrs().MTU
		if ru.MTU > 0 {
			mtu = min(mtu, ru.MTU)
		}
		mtu = min(mtu, hostVethLink.Attrs().MTU)
		return mtu
	}

	slog.InfoContext(ctx, "route update: "+ru.Route.String(), slog.Any("type", ru.Type))
	switch ru.Type {
	case unix.RTM_NEWROUTE:
		l := findLink(ru.LinkIndex)
		if l == nil {
			return fmt.Errorf("link with index %d not found", ru.LinkIndex)
		}
		mtu := getMtu(l)
		slog.InfoContext(ctx, "adding route inside network namespace", slog.Any("dst", ru.Dst.String()), slog.Any("gw", hostIP.String()), slog.Any("mtu", mtu))
		newRoute := netlink.Route{
			LinkIndex: peerLink.Attrs().Index,
			Dst:       ru.Dst,
			Gw:        hostIP,
			MTU:       mtu,
		}
		err := n.sandboxNetlink.RouteAdd(&newRoute)
		if err != nil {
			return err
		}
		return nil
	case unix.RTM_DELROUTE:
		l := findLink(ru.LinkIndex)
		if l == nil {
			return fmt.Errorf("link with index %d not found", ru.LinkIndex)
		}
		slog.InfoContext(ctx, "removing route inside network namespace", slog.Any("dst", ru.Dst.String()), slog.Any("gw", hostIP.String()))
		delRoute := netlink.Route{
			LinkIndex: peerLink.Attrs().Index,
			Dst:       ru.Dst,
			Gw:        hostIP,
			//MTU:       getMtu(l),
		}
		err := n.sandboxNetlink.RouteDel(&delRoute)
		if err != nil {
			return err
		}
	default:
		slog.WarnContext(ctx, "unknown route update type", slog.Any("type", ru.Type))
	}
	return nil
}
