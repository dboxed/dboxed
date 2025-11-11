//go:build linux

package sandbox

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/dboxed/dboxed/pkg/boxspec"
	network2 "github.com/dboxed/dboxed/pkg/runner/network"
	"github.com/opencontainers/runc/libcontainer"
)

type Sandbox struct {
	Debug bool

	HostWorkDir string

	InfraImage  string
	SandboxName string
	SandboxDir  string

	VethNetworkCidr string

	network      *network2.Network
	routesMirror network2.RoutesMirror
}

func (rn *Sandbox) Destroy(ctx context.Context) error {
	c, err := rn.GetSandboxContainer()
	if err != nil {
		if !errors.Is(err, libcontainer.ErrNotExist) && !os.IsNotExist(err) {
			return err
		}
	} else {
		err = rn.StopOrKillSandboxContainer(ctx, time.Second*30, time.Second*10)
		if err != nil {
			return err
		}
		slog.InfoContext(ctx, "destroying old sandbox container")
		err = c.Destroy()
		if err != nil {
			return err
		}
	}

	slog.InfoContext(ctx, "removing sandbox rootfs")
	err = os.RemoveAll(rn.GetSandboxRoot())
	if err != nil {
		return err
	}
	return nil
}

func (rn *Sandbox) Prepare(ctx context.Context) error {
	err := os.MkdirAll(GetContainerStateDir(rn.SandboxDir), 0700)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Join(rn.SandboxDir, "logs"), 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(rn.SandboxDir, "netbird"), 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(rn.SandboxDir, "volumes"), 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(rn.SandboxDir, "containers"), 0700)
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
	err := rn.copyDboxedIntoInfraRoot()
	if err != nil {
		return err
	}

	return nil
}

func (rn *Sandbox) PrepareNetworkingConfig() error {
	networkConfig, err := rn.buildNetworkConfig()
	if err != nil {
		return err
	}
	rn.network = &network2.Network{
		InfraContainerRoot: rn.GetSandboxRoot(),
		Config:             networkConfig,
	}
	err = rn.network.InitNamesAndIPs()
	if err != nil {
		return err
	}
	return nil
}

func (rn *Sandbox) SetupNetworking(ctx context.Context) error {
	err := rn.network.SetupNamespaces(ctx)
	if err != nil {
		return err
	}
	err = rn.network.Setup(ctx)
	if err != nil {
		return err
	}

	err = rn.writeDnsProxyResolvConf()
	if err != nil {
		return err
	}

	rn.routesMirror = network2.RoutesMirror{
		NamesAndIps: rn.network.NamesAndIps,
	}
	err = rn.routesMirror.Start(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (rn *Sandbox) DestroyNetworking(ctx context.Context) error {
	err := rn.network.Destroy(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (rn *Sandbox) Start(ctx context.Context) error {
	err := rn.createAndStartSandboxContainer(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (rn *Sandbox) buildNetworkConfig() (*boxspec.NetworkConfig, error) {
	_, cidr, err := net.ParseCIDR(rn.VethNetworkCidr)
	if err != nil {
		return nil, err
	}
	namesAndIps, err := network2.NewNamesAndIPs(rn.SandboxName, cidr)
	if err != nil {
		return nil, err
	}
	cfg := &boxspec.NetworkConfig{
		SandboxName:     rn.SandboxName,
		VethNetworkCidr: cidr,
		DnsProxyIP:      namesAndIps.PeerAddr.IP.String(),
	}

	return cfg, nil
}
