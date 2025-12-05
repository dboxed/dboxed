//go:build linux

package sandbox

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/dboxed/dboxed/pkg/runner/consts"
	dns_proxy "github.com/dboxed/dboxed/pkg/runner/dns-proxy"
	"github.com/dboxed/dboxed/pkg/runner/logs"
	"github.com/dboxed/dboxed/pkg/runner/network"
)

type RunDnsProxy struct {
}

func (cmd *RunDnsProxy) Run(logHandler *logs.MultiLogHandler) error {
	ctx := context.Background()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(sigs)
	}()

	logFile := filepath.Join(consts.LogsDir, "dns-proxy.log")
	logWriter := logs.BuildRotatingLogger(logFile)

	logHandler.AddWriter(logWriter)
	defer logHandler.RemoveWriter(logWriter)

	hostNetNsFd, err := network.ReadNetNsFD(ctx, consts.NetNsHolderUnixSocket)
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
