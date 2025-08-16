package sandbox

import (
	"context"
	"net"

	dns_proxy "github.com/dboxed/dboxed/pkg/dns-proxy"
)

func (rn *Sandbox) startDnsProxy(ctx context.Context) error {
	rn.dnsProxy = &dns_proxy.DnsProxy{
		ListenNamespace: rn.network.NetworkNamespace,
		QueryNamespace:  rn.network.HostNetworkNamespace,
		ListenIP:        net.ParseIP(rn.network.Config.DnsProxyIP),
		HostFsPath:      "/",
	}

	err := rn.dnsProxy.Start(ctx)
	if err != nil {
		return err
	}

	return nil
}
