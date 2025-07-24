package run_infra_sandbox

import (
	"context"
	"encoding/json"
	"github.com/koobox/unboxed/pkg/dns"
	"github.com/koobox/unboxed/pkg/types"
	"github.com/koobox/unboxed/pkg/util"
	"github.com/vishvananda/netlink"
	"log/slog"
	"time"
)

func (rn *RunInfraSandbox) startDnsPubSub(ctx context.Context) error {
	rn.dnsStore = dns.NewDnsStore(30 * time.Second)

	if rn.conf.BoxSpec.Dns.LibP2P != nil {
		err := rn.startDnsPubSubLibP2P(ctx)
		if err != nil {
			return err
		}
	} else {
		slog.InfoContext(ctx, "no dns publisher configured")
		return nil
	}

	err := rn.watchDnsNetworkInterface(ctx, func(ip string) {
		slog.InfoContext(ctx, "setting dns announcement", slog.Any("ip", ip))
		rn.dnsPubSub.SetDnsAnnouncement(rn.conf.BoxSpec.Dns.Hostname, ip)
	})
	if err != nil {
		return err
	}

	go rn.runWriteStaticHostsMap(ctx)

	return nil
}

func (rn *RunInfraSandbox) startDnsPubSubLibP2P(ctx context.Context) error {
	dnsP2P := &dns.DnsLibP2P{
		Store:         rn.dnsStore,
		NetworkDomain: rn.conf.BoxSpec.Dns.NetworkDomain,
	}
	err := dnsP2P.Start(ctx)
	if err != nil {
		return err
	}
	rn.dnsPubSub = dnsP2P
	return nil
}

func (rn *RunInfraSandbox) watchDnsNetworkInterface(ctx context.Context, fn func(ip string)) error {
	slog.InfoContext(ctx, "subscribing for link updates")
	ch := make(chan netlink.LinkUpdate)
	done := make(chan struct{})
	err := netlink.LinkSubscribeWithOptions(ch, done, netlink.LinkSubscribeOptions{
		ListExisting: true,
	})
	if err != nil {
		return err
	}

	go func() {
		defer close(done)
		for lu := range ch {
			slog.DebugContext(ctx, "link update received", slog.Any("linkName", lu.Attrs().Name))
			if lu.Attrs().Name != rn.conf.BoxSpec.Dns.NetworkInterface {
				continue
			}
			slog.DebugContext(ctx, "link is matching DNS network interface")

			addrs, err := netlink.AddrList(lu.Link, netlink.FAMILY_V4)
			if err != nil {
				slog.ErrorContext(ctx, "error while getting link addresses", slog.Any("error", err))
				continue
			}
			if len(addrs) >= 1 {
				slog.DebugContext(ctx, "notifying about link IP", slog.Any("linkAddrs", addrs))
				ip := addrs[0].IP.String()
				fn(ip)
			}
		}
	}()
	return nil
}

func (rn *RunInfraSandbox) runWriteStaticHostsMap(ctx context.Context) {
	util.LoopWithPrintErr(ctx, "runWriteStaticHostsMapOnce", 5*time.Second, func() error {
		return rn.runWriteStaticHostsMapOnce()
	})
}

func (rn *RunInfraSandbox) runWriteStaticHostsMapOnce() error {
	m := rn.dnsStore.Map()
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	h := util.Sha256Sum(b)
	if h == rn.oldDnsMapHash {
		return nil
	}

	err = util.AtomicWriteFile(types.DnsMapFile, b, 0644)
	if err != nil {
		return err
	}

	rn.oldDnsMapHash = h

	return nil
}
