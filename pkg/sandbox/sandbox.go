package sandbox

import (
	"context"
	"github.com/koobox/unboxed/pkg/types"
	"log/slog"
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
}

func (rn *Sandbox) Start(ctx context.Context) error {
	err := os.MkdirAll(getRuncStateDir(rn.SandboxDir), 0700)
	if err != nil {
		return err
	}
	err = rn.destroyInfraContainer(ctx)
	if err != nil {
		return err
	}
	err = os.RemoveAll(filepath.Join(rn.SandboxDir, "containerd"))
	if err != nil {
		return err
	}

	err = os.MkdirAll(rn.getInfraContainerDir(), 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(rn.SandboxDir, "logs"), 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(rn.SandboxDir, "containerd"), 0700)
	if err != nil {
		return err
	}

	err = rn.pullInfraImage(ctx)
	if err != nil {
		return err
	}

	err = rn.copyRuncFromInfraContainer()
	if err != nil {
		return err
	}
	err = rn.copyUnboxedBinIntoInfraRoot()
	if err != nil {
		return err
	}

	err = rn.writeInfraConf()
	if err != nil {
		return err
	}

	err = os.Symlink("/hostfs/etc/resolv.conf", filepath.Join(rn.getInfraRoot(), "etc/resolv.conf"))
	if err != nil {
		return err
	}

	// TODO move this into the sandbox container when https://github.com/opencontainers/runc/issues/2826 gets fixed
	slog.InfoContext(ctx, "enabling ip forwarding")
	err = os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), 0600)
	if err != nil {
		return err
	}

	err = rn.createInfraContainer(ctx)
	if err != nil {
		return err
	}
	err = rn.startInfraContainer(ctx)
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
