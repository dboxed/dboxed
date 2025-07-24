package run_infra_host

import (
	"context"
	dns_proxy "github.com/koobox/unboxed/pkg/dns-proxy"
	"github.com/koobox/unboxed/pkg/logs"
	"github.com/koobox/unboxed/pkg/network"
	"github.com/koobox/unboxed/pkg/sandbox"
	"github.com/koobox/unboxed/pkg/types"
	"github.com/koobox/unboxed/pkg/util"
	"io"
	"log/slog"
	"net"
	"os"
	"time"
)

type RunInfraHost struct {
	conf *types.InfraConfig

	network         *network.Network
	routesMirror    network.RoutesMirror
	netbirdRulesFix network.NetbirdRulesFix

	dnsProxy      *dns_proxy.DnsProxy
	oldDnsMapHash string

	infraStdout io.WriteCloser
	infraStderr io.WriteCloser
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
		Config: rn.conf.NetworkConfig,
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

func (rn *RunInfraHost) initLogging() {
	infraLog := logs.BuildRotatingLogger("/var/log/unboxed/infra-host.log")
	rn.infraStdout = logs.NewJsonFileLogger(infraLog, "stdout")
	rn.infraStderr = logs.NewJsonFileLogger(infraLog, "stderr")

	handler := slog.NewTextHandler(rn.infraStderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))
}

func (rn *RunInfraHost) stopLogging() {
	_ = rn.infraStderr.Close()
	_ = rn.infraStdout.Close()
}
