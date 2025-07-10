package run_infra_host

import (
	"context"
	dns_proxy "github.com/koobox/unboxed/pkg/dns-proxy"
	"github.com/koobox/unboxed/pkg/network"
	"github.com/koobox/unboxed/pkg/sandbox"
	"github.com/koobox/unboxed/pkg/types"
	"github.com/vishvananda/netns"
	"log/slog"
	"net"
)

type RunInfraHost struct {
	conf *types.InfraConfig

	DnsProxy          *dns_proxy.DnsProxy
	olsStaticHostsMap map[string]string

	routesMirror    network.RoutesMirror
	netbirdRulesFix network.NetbirdRulesFix
}

func (rn *RunInfraHost) Start(ctx context.Context) error {
	slog.InfoContext(ctx, "running infra in host namespace")

	var err error
	rn.conf, err = sandbox.ReadInfraConf(types.InfraConfFile)
	if err != nil {
		return err
	}

	namesAndIps, err := network.NewNamesAndIPs(rn.conf.NetworkConfig)
	if err != nil {
		return err
	}

	hostNamespace, err := netns.Get()
	if err != nil {
		return err
	}
	sandboxNamespace, err := netns.GetFromName(namesAndIps.SandboxNamespaceName)
	if err != nil {
		return err
	}

	rn.routesMirror = network.RoutesMirror{
		NamesAndIps: namesAndIps,
	}
	rn.netbirdRulesFix = network.NetbirdRulesFix{
		SandboxNetworkNamespace: sandboxNamespace,
	}
	err = rn.routesMirror.Start(ctx)
	if err != nil {
		return err
	}
	err = rn.netbirdRulesFix.Start(ctx)
	if err != nil {
		return err
	}

	rn.DnsProxy = &dns_proxy.DnsProxy{
		ListenNamespace: sandboxNamespace,
		QueryNamespace:  hostNamespace,
		ListenIP:        net.ParseIP(rn.conf.NetworkConfig.DnsProxyIP),
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

	go rn.runNetbirdStatusToDns(ctx)

	return nil
}
