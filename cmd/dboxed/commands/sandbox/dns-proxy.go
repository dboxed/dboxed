//go:build linux

package sandbox

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/dboxed/dboxed/pkg/runner/consts"
	dns_proxy "github.com/dboxed/dboxed/pkg/runner/dns-proxy"
	"github.com/dboxed/dboxed/pkg/runner/network"
	"github.com/dboxed/dboxed/pkg/runner/sendnshandle"
)

type RunDnsProxy struct {
}

func (cmd *RunDnsProxy) Run() error {
	ctx := context.Background()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(sigs)
	}()

	hostNetNsFd, err := sendnshandle.ReadNetNsFD(ctx, consts.NetNsHolderUnixSocket)
	if err != nil {
		return err
	}
	defer hostNetNsFd.Close()

	err = network.SetupSandboxDnsProxyIp(ctx, consts.SandboxDnsProxyIp)
	if err != nil {
		return err
	}

	dnsProxy := dns_proxy.DnsProxy{
		ListenIP:             consts.SandboxDnsProxyIp,
		HostResolvConfFile:   consts.HostResolvConfFile,
		HostNetworkNamespace: hostNetNsFd,
	}
	err = dnsProxy.Start(ctx)
	if err != nil {
		return err
	}

	sig := <-sigs
	slog.Info("received signal", "signal", sig.String())

	return nil
}
