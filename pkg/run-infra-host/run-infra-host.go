package run_infra_host

import (
	"context"
	"log/slog"
	"net"
	"os"
	"time"

	dns_proxy "github.com/dboxed/dboxed/pkg/dns-proxy"
	"github.com/dboxed/dboxed/pkg/network"
	"github.com/dboxed/dboxed/pkg/sandbox"
	"github.com/dboxed/dboxed/pkg/types"
	"github.com/dboxed/dboxed/pkg/util"
)

type RunInfraHost struct {
	conf *types.InfraConfig

	network         *network.Network
	routesMirror    network.RoutesMirror
	netbirdRulesFix network.NetbirdRulesFix

	dnsProxy      *dns_proxy.DnsProxy
	oldDnsMapHash string
}

func (rn *RunInfraHost) Run(ctx context.Context) {
	rn.initLogging()
	defer rn.stopLogging()

	err := rn.doRun(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "run-infra-host failed", slog.Any("error", err))
		os.Exit(1)
	}
	os.Exit(0)
}

func (rn *RunInfraHost) doRun(ctx context.Context) error {
	slog.InfoContext(ctx, "running in host container")

	var err error
	rn.conf, err = sandbox.ReadInfraConf(types.InfraConfFile)
	if err != nil {
		return err
	}

	rn.network = &network.Network{
		InfraContainerRoot: "/",
		Config:             rn.conf.NetworkConfig,
	}
	err = rn.network.InitNamesAndIPs()
	if err != nil {
		return err
	}
	err = rn.network.SetupNamespaces(ctx)
	if err != nil {
		return err
	}
	err = rn.network.Setup(ctx)
	if err != nil {
		return err
	}

	rn.routesMirror = network.RoutesMirror{
		NamesAndIps: rn.network.NamesAndIps,
	}
	rn.netbirdRulesFix = network.NetbirdRulesFix{
		SandboxNetworkNamespace: rn.network.NetworkNamespace,
	}
	err = rn.routesMirror.Start(ctx)
	if err != nil {
		return err
	}
	err = rn.netbirdRulesFix.Start(ctx)
	if err != nil {
		return err
	}

	rn.dnsProxy = &dns_proxy.DnsProxy{
		ListenNamespace: rn.network.NetworkNamespace,
		QueryNamespace:  rn.network.HostNetworkNamespace,
		ListenIP:        net.ParseIP(rn.conf.NetworkConfig.DnsProxyIP),
		HostFsPath:      "/hostfs",
	}

	err = rn.dnsProxy.Start(ctx)
	if err != nil {
		return err
	}

	go rn.runReadDnsMap(ctx)

	// let the GC free it up
	rn.conf.BoxSpec.FileBundles = nil

	err = os.WriteFile(types.InfraHostReadyMarkerFile, nil, 0644)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "up and running")
	for {
		if !util.SleepWithContext(ctx, 1*time.Second) {
			break
		}
	}

	return nil
}
