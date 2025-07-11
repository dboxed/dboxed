package run_infra_sandbox

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
	"time"
)

type RunInfraSandbox struct {
	conf *types.InfraConfig

	network         *network.Network
	routesMirror    network.RoutesMirror
	netbirdRulesFix network.NetbirdRulesFix

	dnsProxy          *dns_proxy.DnsProxy
	olsStaticHostsMap map[string]string

	sandboxStdout io.WriteCloser
	sandboxStderr io.WriteCloser
}

func (rn *RunInfraSandbox) Start(ctx context.Context) error {
	rn.initLogging()
	defer rn.stopLogging()

	defer func() {
		slog.InfoContext(ctx, "exiting...")
		time.Sleep(60 * time.Minute)
	}()

	slog.InfoContext(ctx, "running in sandbox container")

	var err error
	rn.conf, err = sandbox.ReadInfraConf(types.InfraConfFile)
	if err != nil {
		return err
	}

	err = rn.writeFileBundles(ctx)
	if err != nil {
		return err
	}

	rn.network = &network.Network{
		Config: rn.conf.NetworkConfig,
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
		HostResolveConf: "/hostfs/etc/resolv.conf",
	}

	err = rn.dnsProxy.Start(ctx)
	if err != nil {
		return err
	}

	err = rn.startContainerd(ctx)
	if err != nil {
		return err
	}

	err = rn.startNetbirdServiceContainer(ctx)
	if err != nil {
		return err
	}
	err = rn.runNetbirdUp(ctx)
	if err != nil {
		return err
	}

	err = util.RunInNetNs(rn.network.NetworkNamespace, func() error {
		return network.WaitForInterface(ctx, "wt0")
	})
	if err != nil {
		return err
	}

	go rn.runNetbirdStatusToDns(ctx)

	err = rn.runComposeUp(ctx)
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

func (rn *RunInfraSandbox) initLogging() {
	sandboxLog := logs.BuildRotatingLogger("/var/log/unboxed/sandbox.log")
	rn.sandboxStdout = logs.NewJsonFileLogger(sandboxLog, "stdout")
	rn.sandboxStderr = logs.NewJsonFileLogger(sandboxLog, "stderr")

	handler := slog.NewTextHandler(rn.sandboxStderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))
}

func (rn *RunInfraSandbox) stopLogging() {
	_ = rn.sandboxStderr.Close()
	_ = rn.sandboxStdout.Close()
}
