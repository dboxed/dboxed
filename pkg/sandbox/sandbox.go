package sandbox

import (
	"context"
	"net"
	"os"
	"path/filepath"

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

	network      *network.Network
	routesMirror network.RoutesMirror
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

	return nil

}

func (rn *Sandbox) CopyBinaries(ctx context.Context) error {
	err := rn.copyRuncFromInfraRoot()
	if err != nil {
		return err
	}

	err = rn.copyDboxedIntoInfraRoot()
	if err != nil {
		return err
	}

	return nil
}

func (rn *Sandbox) SetupNetworking(ctx context.Context) error {
	networkConfig, err := rn.buildNetworkConfig()
	if err != nil {
		return err
	}
	rn.network = &network.Network{
		InfraContainerRoot: rn.GetSandboxRoot(),
		Config:             networkConfig,
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
	err = rn.routesMirror.Start(ctx)
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

func (rn *Sandbox) buildNetworkConfig() (*types.NetworkConfig, error) {
	namesAndIps, err := network.NewNamesAndIPs(rn.SandboxName, rn.VethNetworkCidr)
	if err != nil {
		return nil, err
	}
	cfg := &types.NetworkConfig{
		SandboxName:     rn.SandboxName,
		VethNetworkCidr: rn.VethNetworkCidr,
		DnsProxyIP:      namesAndIps.PeerAddr.IP.String(),
	}

	return cfg, nil
}
