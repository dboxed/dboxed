package network

import (
	"fmt"
	"net"

	"github.com/dboxed/dboxed-common/util"
	"github.com/dboxed/dboxed/pkg/types"
	net2 "github.com/dboxed/dboxed/pkg/util/net"
	"github.com/vishvananda/netlink"
)

type NamesAndIps struct {
	Base                 string
	SandboxNamespaceName string
	VethNameHost         string
	VethNamePeer         string
	HostAddr             netlink.Addr
	PeerAddr             netlink.Addr
}

func NewNamesAndIPs(cfg types.NetworkConfig) (n NamesAndIps, err error) {
	n.Base = n.buildNameBase(cfg)
	n.SandboxNamespaceName = n.Base
	n.VethNameHost = fmt.Sprintf("%s-host", n.Base)
	n.VethNamePeer = fmt.Sprintf("%s-peer", n.Base)

	hostIP, err := net2.GetIndexedIP(cfg.VethNetworkCidr, 0)
	if err != nil {
		return
	}
	peerIP, err := net2.GetIndexedIP(cfg.VethNetworkCidr, 1)
	if err != nil {
		return
	}
	n.HostAddr = netlink.Addr{
		IPNet: &net.IPNet{
			IP:   hostIP,
			Mask: cfg.VethNetworkCidr.Mask,
		},
	}
	n.PeerAddr = netlink.Addr{
		IPNet: &net.IPNet{
			IP:   peerIP,
			Mask: cfg.VethNetworkCidr.Mask,
		},
	}
	return
}

func (n *NamesAndIps) buildNameBase(cfg types.NetworkConfig) string {
	h := util.Sha256Sum([]byte(cfg.SandboxName))
	h = h[:6]

	return fmt.Sprintf("%s-%s", namePrefix, h)
}
