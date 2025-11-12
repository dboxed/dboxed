//go:build linux

package network

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

type PortForwards struct {
	InfraContainerRoot   string
	NamesAndIps          NamesAndIps
	HostNetworkNamespace netns.NsHandle

	portforwardMutex        sync.Mutex
	portForwardsIptablesCnt int
}

func (n *PortForwards) SetupPortForwards(ctx context.Context, pfs []boxspec.PortForward) error {
	n.portforwardMutex.Lock()
	defer n.portforwardMutex.Unlock()

	pfs = append([]boxspec.PortForward{}, pfs...) //clone

	hostNetlink, err := netlink.NewHandleAt(n.HostNetworkNamespace)
	if err != nil {
		return err
	}
	hostLinks, err := hostNetlink.LinkList()
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "hostLinks", "hostLinks", hostLinks)

	n.portForwardsIptablesCnt++
	newChain := fmt.Sprintf("${NAME_BASE}-pf-%d", (n.portForwardsIptablesCnt%2)+1)
	oldChain := fmt.Sprintf("${NAME_BASE}-pf-%d", ((n.portForwardsIptablesCnt+1)%2)+1)

	script := fmt.Sprintf("iptables -t nat -F %s\n", newChain)
	for _, pf := range pfs {
		for _, hostLink := range hostLinks {
			if hostLink.Attrs().Name == n.NamesAndIps.VethNameHost {
				continue
			}

			addrs, err := hostNetlink.AddrList(hostLink, netlink.FAMILY_V4)
			if err != nil {
				return err
			}

			for _, addr := range addrs {
				if addr.IP.IsLoopback() {
					continue
				}

				l := fmt.Sprintf("iptables -t nat -j DNAT -A %s", newChain)
				protocol := pf.Protocol
				if protocol == "" {
					protocol = "tcp"
				}
				l += fmt.Sprintf(" -p %s -m %s", protocol, protocol)

				l += fmt.Sprintf(" -d %s", addr.IP.String())

				dport := fmt.Sprintf("%d:%d", pf.HostFirstPort, pf.HostLastPort)
				l += fmt.Sprintf(" --dport %s", dport)

				dest := n.NamesAndIps.PeerAddr.IP.String()
				dest += fmt.Sprintf(":%d-%d", pf.SandboxPort, pf.SandboxPort+(pf.HostLastPort-pf.HostFirstPort))
				l += fmt.Sprintf(" --to-destination %s", dest)

				script += l + "\n"
			}
		}
	}

	script += fmt.Sprintf("iptables -t nat -A PREROUTING -j %s -m comment --comment ${NAME_BASE}\n", newChain)
	script += fmt.Sprintf("iptables -t nat -D PREROUTING -j %s -m comment --comment ${NAME_BASE} || true\n", oldChain)
	script += fmt.Sprintf("iptables -t nat -F %s\n", oldChain)

	ipt := Iptables{
		InfraContainerRoot: n.InfraContainerRoot,
		NamesAndIps:        n.NamesAndIps,
		Namespace:          n.HostNetworkNamespace,
	}

	return ipt.runIptablesScript(ctx, script)
}
