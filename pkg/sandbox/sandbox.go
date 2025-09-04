package sandbox

import (
	"context"
	"net"
	"os"
	"path/filepath"

	dns_proxy "github.com/dboxed/dboxed/pkg/dns-proxy"
	"github.com/dboxed/dboxed/pkg/network"
	"github.com/dboxed/dboxed/pkg/types"
)

type Sandbox struct {
	Debug bool

	HostWorkDir string

	InfraImage  string
	SandboxName string
	SandboxDir  string

	VethNetworkCidr *net.IPNet

	network         *network.Network
	routesMirror    network.RoutesMirror
	netbirdRulesFix network.NetbirdRulesFix

	dnsProxy      *dns_proxy.DnsProxy
	oldDnsMapHash string
}

func (rn *Sandbox) Destroy(ctx context.Context) error {
	if _, err := os.Stat(getRuncStateDir(rn.SandboxDir)); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		err := rn.destroySandboxContainer(ctx)
		if err != nil {
			return err
		}
	}
	err := os.RemoveAll(filepath.Join(rn.SandboxDir, "docker"))
	if err != nil {
		return err
	}
	err = os.RemoveAll(rn.GetSandboxRoot())
	if err != nil {
		return err
	}
	return nil
}

func (rn *Sandbox) Stop(ctx context.Context) error {
	return rn.killSandboxContainer(ctx)
}

func (rn *Sandbox) Prepare(ctx context.Context) error {
	err := os.MkdirAll(getRuncStateDir(rn.SandboxDir), 0700)
	if err != nil {
		return err
	}

	err = os.MkdirAll(rn.getSandboxContainerDir(), 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(rn.SandboxDir, "logs"), 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(rn.SandboxDir, "volumes"), 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(rn.SandboxDir, "docker"), 0700)
	if err != nil {
		return err
	}

	err = rn.pullInfraImage(ctx)
	if err != nil {
		return err
	}

	err = rn.copyRuncFromInfraRoot()
	if err != nil {
		return err
	}

	return nil
}

func (rn *Sandbox) SetupNetworking(ctx context.Context) error {
	rn.network = &network.Network{
		InfraContainerRoot: rn.GetSandboxRoot(),
		Config:             rn.buildNetworkConfig(),
	}
	err := rn.network.InitNamesAndIPs()
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

	_ = os.Remove(filepath.Join(rn.GetSandboxRoot(), "etc/resolv.conf"))
	err = rn.writeResolvConf(rn.GetSandboxRoot(), rn.network.Config.DnsProxyIP)
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

	err = rn.startDnsProxy(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (rn *Sandbox) Start(ctx context.Context) error {
	err := rn.createSandboxContainer(ctx)
	if err != nil {
		return err
	}
	err = rn.startSandboxContainer(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (rn *Sandbox) buildNetworkConfig() types.NetworkConfig {
	return types.NetworkConfig{
		SandboxName:     rn.SandboxName,
		VethNetworkCidr: rn.VethNetworkCidr,
		DnsProxyIP:      "127.0.0.1", // TODO
	}
}
