package network

import (
	"context"
	"fmt"
	"github.com/koobox/unboxed/pkg/util"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"log/slog"
)

// watchAndUpdateRoutes will watch for routes on the host network namespace and create mirrored routes inside the unboxed
// network namespace. Each such route uses the host veth interface as gateway (NAT). A simpler solution would be to just
// add a single default route, but this would not respect differences in MTUs per host network interface.
func (n *Network) watchAndUpdateRoutes(ctx context.Context, peerLink netlink.Link) error {
	routeUpdateChan := make(chan netlink.RouteUpdate)
	err := netlink.RouteSubscribeWithOptions(routeUpdateChan, ctx.Done(), netlink.RouteSubscribeOptions{
		ListExisting: true,
		ErrorCallback: func(err error) {
			slog.ErrorContext(ctx, err.Error())
		},
	})
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case ru := <-routeUpdateChan:
				err := n.updateRoute(ctx, ru, peerLink)
				if err != nil {
					slog.ErrorContext(ctx, "failed to update route", slog.Any("error", err))
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (n *Network) updateRoute(ctx context.Context, ru netlink.RouteUpdate, peerLink netlink.Link) error {
	if ru.Dst != nil {
		if ru.Dst.IP.IsLoopback() {
			return nil
		}
		if n.Config.VethNetworkCidr.Contains(ru.Dst.IP) {
			return nil
		}
		if ru.Dst.IP.To4() == nil {
			return nil
		}
	}
	if ru.Src != nil {
		if n.Config.VethNetworkCidr.Contains(ru.Dst.IP) {
			return nil
		}
	}

	hostIP := n.HostAddr.IP

	hostLinks, err := netlink.LinkList()
	if err != nil {
		return err
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
		if ru.MTU > 0 && ru.MTU < mtu {
			mtu = ru.MTU
		}
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
		err := util.RunInNetNs(n.NetworkNamespace, func() error {
			return netlink.RouteAdd(&newRoute)
		})
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
		err := util.RunInNetNs(n.NetworkNamespace, func() error {
			return netlink.RouteDel(&delRoute)
		})
		if err != nil {
			return err
		}
	default:
		slog.WarnContext(ctx, "unknown route update type", slog.Any("type", ru.Type))
	}
	return nil
}
