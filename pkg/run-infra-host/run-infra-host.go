package run_infra_host

import (
	"context"
	dns_proxy "github.com/koobox/unboxed/pkg/dns-proxy"
	"github.com/koobox/unboxed/pkg/sandbox"
	"github.com/koobox/unboxed/pkg/types"
	"github.com/vishvananda/netns"
	"log/slog"
	"net"
)

type RunInfraHost struct {
	conf *types.InfraConfig

	DnsProxy            *dns_proxy.DnsProxy
	staticHostsMapBytes []byte
}

func (rn *RunInfraHost) Start(ctx context.Context) error {
	slog.InfoContext(ctx, "running infra in host namespace")

	var err error
	rn.conf, err = sandbox.ReadInfraConf(types.InfraConfFile)
	if err != nil {
		return err
	}

	hostNamespace, err := netns.Get()
	if err != nil {
		return err
	}
	sandboxNamespace, err := netns.GetFromName(rn.conf.NetworkNamespaceName)
	if err != nil {
		return err
	}

	rn.DnsProxy = &dns_proxy.DnsProxy{
		ListenNamespace: sandboxNamespace,
		QueryNamespace:  hostNamespace,
		ListenIP:        net.ParseIP(rn.conf.DnsProxyIP),
		HostResolveConf: "/hostfs/etc/resolv.conf",
	}

	err = rn.DnsProxy.Start(ctx)
	if err != nil {
		return err
	}

	err = rn.waitForNetbirdContainer(ctx)
	if err != nil {
		return err
	}

	err = rn.runNetbirdUp(ctx)
	if err != nil {
		return err
	}

	go rn.runNetbirdStatusLoop(ctx)
	go rn.runSerfStaticHosts(ctx)

	return nil
}
