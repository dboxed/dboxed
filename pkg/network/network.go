package network

import (
	"fmt"
	"github.com/koobox/unboxed/pkg/types"
	"github.com/koobox/unboxed/pkg/util"
	net2 "github.com/koobox/unboxed/pkg/util/net"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"net"
	"sync"
)

const namePrefix = "ub"

type Network struct {
	Config types.NetworkConfig

	InfraContainerRoot string

	NetworkNamespaceName string
	vethNameHost         string
	vethNamePeer         string
	HostAddr             netlink.Addr
	PeerAddr             netlink.Addr

	HostNetworkNamespace netns.NsHandle
	NetworkNamespace     netns.NsHandle

	portforwardMutex        sync.Mutex
	portForwardsIptablesCnt int
}

func (n *Network) initNamesAndIPs() error {
	base := n.buildNameBase()
	n.NetworkNamespaceName = base
	n.vethNameHost = fmt.Sprintf("%s-host", base)
	n.vethNamePeer = fmt.Sprintf("%s-peer", base)

	hostIP, err := net2.GetIndexedIP(n.Config.VethNetworkCidr, 0)
	if err != nil {
		return err
	}
	peerIP, err := net2.GetIndexedIP(n.Config.VethNetworkCidr, 1)
	if err != nil {
		return err
	}
	n.HostAddr = netlink.Addr{
		IPNet: &net.IPNet{
			IP:   hostIP,
			Mask: n.Config.VethNetworkCidr.Mask,
		},
	}
	n.PeerAddr = netlink.Addr{
		IPNet: &net.IPNet{
			IP:   peerIP,
			Mask: n.Config.VethNetworkCidr.Mask,
		},
	}
	return nil
}

func (n *Network) buildNameBase() string {
	h := util.Sha256Sum([]byte(n.Config.SandboxName))
	h = h[:6]

	return fmt.Sprintf("%s-%s", namePrefix, h)
}
