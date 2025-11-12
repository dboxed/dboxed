package run_in_sandbox

import (
	"context"
	"net"

	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/runner/dns-proxy"
)

func (rn *RunInSandbox) startDnsProxy(ctx context.Context) error {
	rn.dnsProxy = &dns_proxy.DnsProxy{
		ListenIP:           net.ParseIP(consts.SandboxDnsProxyIp),
		HostResolvConfFile: consts.HostResolvConfFile,
	}

	err := rn.dnsProxy.Start(ctx)
	if err != nil {
		return err
	}

	return nil
}
