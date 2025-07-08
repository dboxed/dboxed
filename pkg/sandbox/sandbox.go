package sandbox

import (
	"context"
	"fmt"
	"github.com/koobox/unboxed/pkg/dns-proxy"
	"github.com/koobox/unboxed/pkg/logs"
	"github.com/koobox/unboxed/pkg/network"
	"github.com/koobox/unboxed/pkg/types"
	net2 "github.com/koobox/unboxed/pkg/util/net"
	"github.com/koobox/unboxed/pkg/version"
	"net"
	"os"
	"path/filepath"
)

type Sandbox struct {
	HostWorkDir string

	SandboxName string
	SandboxDir  string

	BoxSpec *types.BoxSpec

	netbirdContainerSpec *types.ContainerSpec
	infraContainerSpec   *types.ContainerSpec

	VethNetworkCidr *net.IPNet
	network         *network.Network

	DnsProxy            *dns_proxy.DnsProxy
	staticHostsMapBytes []byte
}

func (rn *Sandbox) Start(ctx context.Context) error {
	err := os.MkdirAll(filepath.Join(rn.SandboxDir, "containers"), 0700)
	if err != nil {
		return err
	}

	err = rn.destroyContainers(ctx)
	if err != nil {
		return err
	}

	rn.infraContainerSpec = rn.buildInfraContainerSpec()
	rn.netbirdContainerSpec, err = rn.buildNetbirdContainerSpec()
	if err != nil {
		return err
	}

	err = rn.pullImages(ctx)
	if err != nil {
		return err
	}

	networkConfig := rn.buildNetworkConfig()
	rn.network = &network.Network{
		Config:             networkConfig,
		InfraContainerRoot: rn.getContainerRoot("_infra"),
	}

	err = rn.network.Setup(ctx)
	if err != nil {
		return err
	}

	// we mount the host log dir into the infra container, so we need to create it first
	err = os.MkdirAll(filepath.Join(rn.getContainerRoot("_infra"), logs.RootLogDir[1:]), 0700)
	if err != nil {
		return err
	}
	// we also need it on the host
	err = os.MkdirAll(logs.RootLogDir, 0700)
	if err != nil {
		return err
	}

	err = rn.copyRuncFromInfraContainer()
	if err != nil {
		return err
	}
	err = rn.copyUnboxedBinIntoInfraContainer()
	if err != nil {
		return err
	}
	err = rn.copyBoxSpecIntoInfraContainer()
	if err != nil {
		return err
	}
	err = rn.forAllContainers(func(c *types.ContainerSpec) error {
		return rn.writeMiscFiles(c)
	})

	dnsListenIp, err := net2.GetIndexedIP(rn.VethNetworkCidr, 1)
	if err != nil {
		return err
	}
	rn.DnsProxy = &dns_proxy.DnsProxy{
		ListenNamespace: rn.network.NetworkNamespace,
		QueryNamespace:  rn.network.HostNetworkNamespace,
		ListenIP:        dnsListenIp,
	}

	err = rn.forAllContainers(func(c *types.ContainerSpec) error {
		return rn.DnsProxy.WriteResolvConf(rn.getContainerRoot(c.Name))
	})
	if err != nil {
		return err
	}

	err = rn.DnsProxy.Start(ctx)
	if err != nil {
		return err
	}

	err = rn.forAllContainers(func(c *types.ContainerSpec) error {
		return rn.createContainer(ctx, c)
	})
	if err != nil {
		return err
	}
	err = rn.forAllContainers(func(c *types.ContainerSpec) error {
		return rn.startContainer(ctx, c)
	})
	if err != nil {
		return err
	}

	err = rn.runNetbirdUp(ctx)
	if err != nil {
		return err
	}

	go rn.runNetbirdStatusLoop(ctx)
	go rn.runSerfStaticHosts(ctx)

	return nil
}

func (rn *Sandbox) buildNetworkConfig() types.NetworkConfig {
	return types.NetworkConfig{
		SandboxName:     rn.SandboxName,
		VethNetworkCidr: rn.VethNetworkCidr,
	}
}

func (rn *Sandbox) buildInfraContainerSpec() *types.ContainerSpec {
	infraImage := rn.BoxSpec.InfraImage
	if infraImage == "" {
		tag := "nightly"
		if !version.IsDummyVersion() {
			tag = version.Version
		}
		infraImage = fmt.Sprintf("%s:%s", types.UnboxedInfraImage, tag)
	}

	return &types.ContainerSpec{
		Name:  "_infra",
		Image: infraImage,
		Cmd: []string{
			"/usr/bin/unboxed",
			"run-infra",
		},
		Privileged:  true,
		UseDevTmpFs: true,
	}
}

func (rn *Sandbox) forAllContainers(fn func(c *types.ContainerSpec) error) error {
	err := fn(rn.netbirdContainerSpec)
	if err != nil {
		return err
	}
	err = fn(rn.infraContainerSpec)
	if err != nil {
		return err
	}
	for _, c := range rn.BoxSpec.Containers {
		err := fn(&c)
		if err != nil {
			return err
		}
	}
	return nil
}
