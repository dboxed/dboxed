//go:build linux

package network

import (
	"fmt"
	"net"

	"github.com/dboxed/dboxed-common/util"
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

func NewNamesAndIPs(sandboxName string, vethCidr *net.IPNet) (n NamesAndIps, err error) {
	n.Base = n.buildNameBase(sandboxName)
	n.SandboxNamespaceName = n.Base
	n.VethNameHost = fmt.Sprintf("%s-host", n.Base)
	n.VethNamePeer = fmt.Sprintf("%s-peer", n.Base)

	hostIP, err := net2.GetIndexedIP(vethCidr, 0)
	if err != nil {
		return
	}
	peerIP, err := net2.GetIndexedIP(vethCidr, 1)
	if err != nil {
		return
	}
	n.HostAddr = netlink.Addr{
		IPNet: &net.IPNet{
			IP:   hostIP,
			Mask: vethCidr.Mask,
		},
	}
	n.PeerAddr = netlink.Addr{
		IPNet: &net.IPNet{
			IP:   peerIP,
			Mask: vethCidr.Mask,
		},
	}
	return
}

func (n *NamesAndIps) buildNameBase(sandboxName string) string {
	h := util.Sha256Sum([]byte(sandboxName))
	h = h[:6]

	return fmt.Sprintf("%s-%s", namePrefix, h)
}
