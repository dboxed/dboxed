//go:build linux

package network

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/vishvananda/netlink"
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

func setSingleAddress(ctx context.Context, netlinkHandle *netlink.Handle, link netlink.Link, addr netlink.Addr) error {
	addrs, err := netlinkHandle.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		return err
	}
	if !slices.ContainsFunc(addrs, func(addr netlink.Addr) bool {
		return addr.Equal(addr)
	}) {
		if len(addrs) != 0 {
			return fmt.Errorf("link %s already contains an IP address that does not belong to dboxed", link.Attrs().Name)
		}

		slog.InfoContext(ctx, fmt.Sprintf("adding address %s to link %s", addr.String(), link.Attrs().Name))
		err = netlinkHandle.AddrAdd(link, &addr)
		if err != nil {
			return err
		}
	} else {
		slog.InfoContext(ctx, fmt.Sprintf("link %s already has address %s", link.Attrs().Name, addr.String()))
	}
	return nil
}
