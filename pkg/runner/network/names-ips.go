//go:build linux

package network

import (
	"fmt"
	"net"

	"github.com/dboxed/dboxed/pkg/runner/consts"
	"github.com/dboxed/dboxed/pkg/util"
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

func NewNamesAndIPs(sandboxId string, vethCidrStr string) (n NamesAndIps, err error) {
	_, vethCidr, err := net.ParseCIDR(vethCidrStr)
	if err != nil {
		return NamesAndIps{}, err
	}

	n.Base = n.buildNameBase(sandboxId)
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

func (n *NamesAndIps) buildNameBase(sandboxId string) string {
	h := util.Sha256Sum([]byte(sandboxId))
	h = h[:6]

	return fmt.Sprintf("%s-%s", consts.SandboxShortPrefix, h)
}
