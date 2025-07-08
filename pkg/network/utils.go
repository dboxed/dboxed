package network

import (
	"context"
	"fmt"
	"github.com/koobox/unboxed/pkg/util"
	"github.com/vishvananda/netlink"
	"log/slog"
	"net"
	"slices"
	"time"
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

func WaitForInterface(ctx context.Context, name string) error {
	slog.InfoContext(ctx, fmt.Sprintf("waiting for %s to come up", name))

	for {
		l, err := netlink.LinkByName(name)
		if err != nil {
			if !isLinkNotFoundError(err) {
				return err
			}
		} else {
			if l.Attrs().Flags&net.FlagUp != 0 {
				slog.InfoContext(ctx, fmt.Sprintf("%s is up", name))
				return nil
			}
		}

		if !util.SleepWithContext(ctx, time.Second) {
			return ctx.Err()
		}
	}
}
