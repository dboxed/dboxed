//go:build linux

package network

import (
	"context"
	"fmt"
	"sync"

	"github.com/dboxed/dboxed/pkg/boxspec"
)

type PortForwards struct {
	InfraContainerRoot string
	NamesAndIps        NamesAndIps

	portforwardMutex        sync.Mutex
	portForwardsIptablesCnt int
}

func (n *PortForwards) SetupPortForwards(ctx context.Context, pfs []boxspec.PortForward) error {
	n.portforwardMutex.Lock()
	defer n.portforwardMutex.Unlock()

	pfs = append([]boxspec.PortForward{}, pfs...) //clone

	n.portForwardsIptablesCnt++
	newChain := fmt.Sprintf("${NAME_PREFIX}-%d", (n.portForwardsIptablesCnt%2)+1)
	oldChain := fmt.Sprintf("${NAME_PREFIX}-%d", ((n.portForwardsIptablesCnt+1)%2)+1)

	script := fmt.Sprintf("$IPTABLES -t nat -F %s\n", newChain)
	for _, pf := range pfs {
		l := fmt.Sprintf("$IPTABLES -t nat -j DNAT -A %s", newChain)
		protocol := pf.Protocol
		if protocol == "" {
			protocol = "tcp"
		}
		l += fmt.Sprintf(" -p %s -m %s", protocol, protocol)

		if pf.IP != "" {
			l += fmt.Sprintf(" -d %s", pf.IP)
		}

		dport := fmt.Sprintf("%d:%d", pf.HostFirstPort, pf.HostLastPort)
		l += fmt.Sprintf(" --dport %s", dport)

		dest := n.NamesAndIps.PeerAddr.IP.String()
		dest += fmt.Sprintf(":%d-%d", pf.SandboxPort, pf.SandboxPort+(pf.HostLastPort-pf.HostFirstPort))
		l += fmt.Sprintf(" --to-destination %s", dest)

		script += l + "\n"
	}

	script += fmt.Sprintf("$IPTABLES -t nat -A PREROUTING -j %s -m comment --comment ${NAME_PREFIX}\n", newChain)
	script += fmt.Sprintf("$IPTABLES -t nat -D PREROUTING -j %s -m comment --comment ${NAME_PREFIX} || true\n", oldChain)
	script += fmt.Sprintf("$IPTABLES -t nat -F %s\n", oldChain)

	ipt := Iptables{
		InfraContainerRoot: n.InfraContainerRoot,
		NamesAndIps:        n.NamesAndIps,
	}

	return ipt.runIptablesScript(ctx, script)
}
