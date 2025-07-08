package network

import (
	"context"
	"fmt"
	"github.com/vishvananda/netlink"
	"log/slog"
	"slices"
)

func isLinkNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(netlink.LinkNotFoundError); ok {
		return true
	}
	return false
}

func setSingleAddress(ctx context.Context, link netlink.Link, addr netlink.Addr) error {
	addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		return err
	}
	if !slices.ContainsFunc(addrs, func(addr netlink.Addr) bool {
		return addr.Equal(addr)
	}) {
		if len(addrs) != 0 {
			return fmt.Errorf("link %s already contains an IP address that does not belong to unboxed", link.Attrs().Name)
		}

		slog.InfoContext(ctx, fmt.Sprintf("adding address %s to link %s", addr.String(), link.Attrs().Name))
		err = netlink.AddrAdd(link, &addr)
		if err != nil {
			return err
		}
	}
	return nil
}
