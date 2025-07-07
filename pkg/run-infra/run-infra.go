package run_infra

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/serf/serf"
	"github.com/koobox/unboxed/pkg/types"
	"github.com/koobox/unboxed/pkg/util"
	"github.com/vishvananda/netlink"
	"log/slog"
	"net"
	"os"
	"time"
)

type RunInfra struct {
	boxSpec types.BoxSpec

	serf                   *serf.Serf
	serfJoinedIps          map[string]struct{}
	serfMembersFileContent []byte
}

func (rn *RunInfra) Start(ctx context.Context) error {
	slog.InfoContext(ctx, "running infra")

	boxSpecBytes, err := os.ReadFile("/etc/unboxed/box-spec.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(boxSpecBytes, &rn.boxSpec)
	if err != nil {
		return err
	}

	err = rn.waitForWireguardInterface(ctx)
	if err != nil {
		return err
	}

	err = rn.startSerf(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (rn *RunInfra) waitForWireguardInterface(ctx context.Context) error {
	ifaceName := "wt0"

	slog.InfoContext(ctx, fmt.Sprintf("waiting for %s to come up", ifaceName))

	for {
		l, err := netlink.LinkByName(ifaceName)
		if err != nil {
			if _, ok := err.(netlink.LinkNotFoundError); !ok {
				return err
			}
		} else {
			if l.Attrs().Flags&net.FlagUp != 0 {
				slog.InfoContext(ctx, fmt.Sprintf("%s is up", ifaceName))
				return nil
			}
		}

		if !util.SleepWithContext(ctx, time.Second) {
			return ctx.Err()
		}
	}
}
