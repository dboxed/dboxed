package sandbox

import (
	"context"
	"github.com/koobox/unboxed/pkg/dns-proxy"
	"github.com/koobox/unboxed/pkg/logs"
	"github.com/koobox/unboxed/pkg/types"
	net2 "github.com/koobox/unboxed/pkg/util/net"
	"github.com/vishvananda/netns"
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

	HostNetworkNamespace netns.NsHandle

	NetworkNamespaceName string
	NetworkNamespace     netns.NsHandle

	PortForwardsIptablesCnt int
	PortForwardsHash        string

	vethNameHost string
	vethNamePeer string

	DnsProxy            *dns_proxy.DnsProxy
	staticHostsMapBytes []byte
}

func (rn *Sandbox) Start(ctx context.Context) error {
	rn.initNetworkNames()

	err := os.MkdirAll(filepath.Join(rn.SandboxDir, "containers"), 0700)
	if err != nil {
		return err
	}

	err = rn.destroyContainers(ctx)
	if err != nil {
		return err
	}
	err = rn.destroyNetworking(ctx)
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

	err = rn.setupNetworking(ctx)
	if err != nil {
		return err
	}

	dnsListenIp, err := net2.GetIndexedIP(rn.VethNetworkCidr, 1)
	if err != nil {
		return err
	}
	rn.DnsProxy = &dns_proxy.DnsProxy{
		ListenNamespace: rn.NetworkNamespace,
		QueryNamespace:  rn.HostNetworkNamespace,
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

func (rn *Sandbox) buildInfraContainerSpec() *types.ContainerSpec {
	return &types.ContainerSpec{
		Name:  "_infra",
		Image: rn.BoxSpec.InfraImage,
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
