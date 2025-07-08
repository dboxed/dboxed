package network

import (
	"context"
	"fmt"
	"github.com/koobox/unboxed/pkg/types"
)

func (n *Network) SetupPortForwards(ctx context.Context, pfs []types.PortForward) error {
	n.portforwardMutex.Lock()
	defer n.portforwardMutex.Unlock()

	pfs = append([]types.PortForward{}, pfs...) //clone
	pfs = append(pfs, types.PortForward{
		Protocol: "udp",
		FromPort: 51820,
	})

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

		dport := fmt.Sprintf("%d", pf.FromPort)
		if pf.ToPort != 0 {
			dport += fmt.Sprintf(":%d", pf.ToPort)
		}
		l += fmt.Sprintf(" --dport %s", dport)

		dest := n.peerAddr.IP.String()
		if pf.DestinationPort != 0 {
			dest += fmt.Sprintf(":%d", pf.DestinationPort)
		}
		l += fmt.Sprintf(" --to-destination %s", dest)

		script += l + "\n"
	}

	script += fmt.Sprintf("$IPTABLES -t nat -A PREROUTING -j %s -m comment --comment ${NAME_PREFIX}\n", newChain)
	script += fmt.Sprintf("$IPTABLES -t nat -D PREROUTING -j %s -m comment --comment ${NAME_PREFIX} || true\n", oldChain)
	script += fmt.Sprintf("$IPTABLES -t nat -F %s\n", oldChain)

	return n.runIptablesScript(ctx, script)
}
