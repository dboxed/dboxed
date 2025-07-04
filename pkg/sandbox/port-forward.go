package sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/koobox/unboxed/pkg/util"
	"github.com/vishvananda/netlink"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"time"
)

type PortForward struct {
	IP string `json:"ip"`

	FromPort int    `json:"fromPort"`
	ToPort   int    `json:"toPort"`
	Protocol string `json:"protocol"`

	DestinationPort int `json:"destinationPort"`
}

type PortForwardsFile struct {
	PortForwards []PortForward `json:"portForwards"`
}

func (rn *Sandbox) watchAndUpdatePortForwards(ctx context.Context, hostAddr netlink.Addr, peerIP net.IP) {
	go func() {
		err := rn.updatePortForwardIptablesRules(ctx, peerIP)
		if err != nil {
			slog.ErrorContext(ctx, "error in doWatchAndUpdatePortForwards", slog.Any("error", err))
		}
		if !util.SleepWithContext(ctx, 5*time.Second) {
			return
		}
	}()
}

func (rn *Sandbox) updatePortForwardIptablesRules(ctx context.Context, peerIP net.IP) error {
	var pff PortForwardsFile
	f, err := os.ReadFile(filepath.Join(rn.SandboxDir, "port-forwards.json"))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	h := util.Sha256Sum(f)
	if h == rn.PortForwardsHash {
		return nil
	}
	rn.PortForwardsHash = h

	err = json.Unmarshal(f, &pff)
	if err != nil {
		return err
	}

	// wireguard/netbird
	pff.PortForwards = append(pff.PortForwards, PortForward{
		Protocol: "udp",
		FromPort: 51820,
	})

	rn.PortForwardsIptablesCnt++
	newChain := fmt.Sprintf("${NAME_PREFIX}-%d", (rn.PortForwardsIptablesCnt%2)+1)
	oldChain := fmt.Sprintf("${NAME_PREFIX}-%d", ((rn.PortForwardsIptablesCnt+1)%2)+1)

	script := fmt.Sprintf("$IPTABLES -t nat -F %s\n", newChain)
	for _, pf := range pff.PortForwards {
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

		dest := peerIP.String()
		if pf.DestinationPort != 0 {
			dest += fmt.Sprintf(":%d", pf.DestinationPort)
		}
		l += fmt.Sprintf(" --to-destination %s", dest)

		script += l + "\n"
	}

	script += fmt.Sprintf("$IPTABLES -t nat -A PREROUTING -j %s -m comment --comment ${NAME_PREFIX}\n", newChain)
	script += fmt.Sprintf("$IPTABLES -t nat -D PREROUTING -j %s -m comment --comment ${NAME_PREFIX} || true\n", oldChain)
	script += fmt.Sprintf("$IPTABLES -t nat -F %s\n", oldChain)

	return rn.runIptablesScript(ctx, script)
}
