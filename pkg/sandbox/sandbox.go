package sandbox

import (
	"context"
	"github.com/koobox/unboxed/pkg/network"
	"github.com/koobox/unboxed/pkg/types"
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

	VethNetworkCidr *net.IPNet

	network *network.Network
}

func (rn *Sandbox) Start(ctx context.Context) error {
	err := os.MkdirAll(getRuncStateDir(rn.SandboxDir), 0700)
	if err != nil {
		return err
	}
	err = rn.destroyInfraContainer(ctx, "infra-sandbox")
	if err != nil {
		return err
	}
	err = rn.destroyInfraContainer(ctx, "infra-host")
	if err != nil {
		return err
	}
	err = os.RemoveAll(filepath.Join(rn.SandboxDir, "docker"))
	if err != nil {
		return err
	}
	err = os.RemoveAll(rn.getInfraRoot())
	if err != nil {
		return err
	}

	err = os.MkdirAll(rn.getInfraContainerDir("infra-host"), 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(rn.getInfraContainerDir("infra-sandbox"), 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(rn.SandboxDir, "logs"), 0700)
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
	err = rn.copyUnboxedBinIntoInfraRoot()
	if err != nil {
		return err
	}

	rn.network = &network.Network{
		Config: rn.buildNetworkConfig(),
	}
	err = rn.network.InitNamesAndIPs()
	if err != nil {
		return err
	}
	err = rn.network.SetupNamespaces(ctx)
	if err != nil {
		return err
	}

	err = rn.writeInfraConf()
	if err != nil {
		return err
	}

	_ = os.Remove(filepath.Join(rn.getInfraRoot(), "etc/resolv.conf"))
	err = rn.writeResolvConf(rn.getInfraRoot(), rn.network.Config.DnsProxyIP)
	if err != nil {
		return err
	}

	err = rn.createInfraContainer(ctx, true, "infra-host", []string{"unboxed", "run-infra-host"})
	if err != nil {
		return err
	}
	err = rn.createInfraContainer(ctx, false, "infra-sandbox", []string{"unboxed", "run-infra-sandbox"})
	if err != nil {
		return err
	}
	err = rn.startInfraContainer(ctx, "infra-host")
	if err != nil {
		return err
	}
	err = rn.startInfraContainer(ctx, "infra-sandbox")
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
