package run_in_sandbox

import (
	"context"
	"fmt"
	"net"

	"github.com/dboxed/dboxed/pkg/runner/dns-proxy"
	"github.com/dboxed/dboxed/pkg/util"
)

func (rn *RunInSandbox) writeResolvConf() error {
	resolveConf := fmt.Sprintf(`# This is the dboxed dns proxy, which listens inside the sandboxed network namespace
# and forwards requests to the host's resolv.conf nameservers
nameserver %s
search .
`, rn.networkConfig.DnsProxyIP)

	err := util.AtomicWriteFile("/etc/resolv.conf", []byte(resolveConf), 0644)
	if err != nil {
		return err
	}
	return nil
}

func (rn *RunInSandbox) startDnsProxy(ctx context.Context) error {
	err := rn.writeResolvConf()
	if err != nil {
		return err
	}

	rn.dnsProxy = &dns_proxy.DnsProxy{
		ListenIP:   net.ParseIP(rn.networkConfig.DnsProxyIP),
		HostFsPath: "/hostfs",
	}

	err = rn.dnsProxy.Start(ctx)
	if err != nil {
		return err
	}

	return nil
}
