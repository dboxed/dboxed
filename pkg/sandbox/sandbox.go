package sandbox

import (
	"context"
	"fmt"
	"github.com/koobox/unboxed/pkg/network"
	"github.com/koobox/unboxed/pkg/types"
	"github.com/koobox/unboxed/pkg/version"
	"net"
	"os"
	"path/filepath"
)

type Sandbox struct {
	Debug bool

	HostWorkDir string

	SandboxName string
	SandboxDir  string

	BoxSpec *types.BoxSpec

	netbirdContainerSpec      types.ContainerSpec
	infraHostContainerSpec    types.ContainerSpec
	infraSandboxContainerSpec types.ContainerSpec

	VethNetworkCidr *net.IPNet
	network         *network.Network
}

func (rn *Sandbox) Start(ctx context.Context) error {
	err := os.MkdirAll(filepath.Join(rn.SandboxDir, "containers"), 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(rn.getSharedDirOnHost(), 0700)
	if err != nil {
		return err
	}

	err = rn.destroyContainers(ctx)
	if err != nil {
		return err
	}

	rn.infraHostContainerSpec = rn.buildInfraHostContainerSpec()
	rn.infraSandboxContainerSpec = rn.buildInfraSandboxContainerSpec()
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
		InfraContainerRoot: rn.getContainerRoot("_infra_host"),
	}

	err = rn.network.Setup(ctx)
	if err != nil {
		return err
	}

	err = rn.copyRuncFromInfraContainer()
	if err != nil {
		return err
	}
	err = rn.forInternalContainers(func(c *types.ContainerSpec) error {
		err := os.MkdirAll(filepath.Join(rn.getContainerRoot(c.Name), types.SharedDir), 0700)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = rn.forAllContainers(func(c *types.ContainerSpec) error {
		err := rn.copyUnboxedBinIntoContainer(c.Name)
		if err != nil {
			return err
		}
		err = rn.writeMiscFiles(c)
		if err != nil {
			return err
		}
		err = rn.writeResolvConf(rn.getContainerRoot(c.Name), rn.network.Config.DnsProxyIP)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = rn.writeInfraConf()
	if err != nil {
		return err
	}

	err = rn.writeFileBundles(ctx)
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

	return nil
}

func (rn *Sandbox) buildNetworkConfig() types.NetworkConfig {
	return types.NetworkConfig{
		SandboxName:     rn.SandboxName,
		VethNetworkCidr: rn.VethNetworkCidr,
		DnsProxyIP:      "127.0.0.1", // TODO
	}
}

func (rn *Sandbox) getInfraImage() string {
	infraImage := rn.BoxSpec.InfraImage
	if infraImage == "" {
		tag := "nightly"
		if !version.IsDummyVersion() {
			tag = version.Version
		}
		infraImage = fmt.Sprintf("%s:%s", types.UnboxedInfraImage, tag)
	}
	return infraImage
}

func (rn *Sandbox) buildInfraHostContainerSpec() types.ContainerSpec {
	return types.ContainerSpec{
		Name:  "_infra_host",
		Image: rn.getInfraImage(),
		Cmd: []string{
			"/usr/bin/unboxed",
			"run-infra-host",
		},
		BindMounts: []types.BindMount{
			{HostPath: "/", ContainerPath: "/hostfs", Shared: true},
			{HostPath: "/run/netns", ContainerPath: "/run/netns", Shared: true},
			{HostPath: rn.getSharedDirOnHost(), ContainerPath: types.SharedDir},
			{HostPath: rn.SandboxDir, ContainerPath: rn.SandboxDir, Shared: true},
		},
		Privileged:  true,
		UseDevTmpFs: true,
		HostNetwork: true,
		HostPid:     true,
		HostCgroups: true,
	}
}

func (rn *Sandbox) buildInfraSandboxContainerSpec() types.ContainerSpec {
	return types.ContainerSpec{
		Name:  "_infra_sandbox",
		Image: rn.getInfraImage(),
		Cmd: []string{
			"/usr/bin/unboxed",
			"run-infra-sandbox",
		},
		BindMounts: []types.BindMount{
			{HostPath: rn.getSharedDirOnHost(), ContainerPath: types.SharedDir},
		},
		Privileged:  true,
		UseDevTmpFs: true,
	}
}

func (rn *Sandbox) internalContainers() []types.ContainerSpec {
	return []types.ContainerSpec{
		rn.infraHostContainerSpec,
		rn.infraSandboxContainerSpec,
		rn.netbirdContainerSpec,
	}
}

func (rn *Sandbox) forInternalContainers(fn func(c *types.ContainerSpec) error) error {
	return rn.forContainers(rn.internalContainers(), fn)
}

func (rn *Sandbox) forAllContainers(fn func(c *types.ContainerSpec) error) error {
	all := rn.internalContainers()
	all = append(all, rn.BoxSpec.Containers...)
	return rn.forContainers(all, fn)
}

func (rn *Sandbox) forContainers(containers []types.ContainerSpec, fn func(c *types.ContainerSpec) error) error {
	for _, c := range containers {
		err := fn(&c)
		if err != nil {
			return err
		}
	}
	return nil
}
