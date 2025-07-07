package run_infra

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/hashicorp/serf/serf"
	"github.com/koobox/unboxed/pkg/types"
	"github.com/koobox/unboxed/pkg/util"
	"log/slog"
	"os"
	"time"
)

func (rn *RunInfra) startSerf(ctx context.Context) error {
	rn.serfJoinedIps = map[string]struct{}{}

	serfConfig := serf.DefaultConfig()
	var err error
	rn.serf, err = serf.Create(serfConfig)
	if err != nil {
		return err
	}

	go func() {
		for {
			err := rn.runSerfJoin(ctx)
			if err != nil {
				slog.ErrorContext(ctx, "error in runSerfJoin", slog.Any("error", err))
			}
			err = rn.writeSerfMembersList(ctx)
			if err != nil {
				slog.ErrorContext(ctx, "error in writeSerfMembersList", slog.Any("error", err))
			}
			if !util.SleepWithContext(ctx, time.Second*5) {
				return
			}
		}
	}()

	return nil
}

func (rn *RunInfra) runSerfJoin(ctx context.Context) error {
	ips, err := rn.getNetbirdPeerIps()
	if err != nil {
		return err
	}

	newMap := map[string]struct{}{}

	var newIps []string
	for _, ip := range ips {
		newMap[ip] = struct{}{}
		if _, ok := rn.serfJoinedIps[ip]; !ok {
			rn.serfJoinedIps[ip] = struct{}{}
			newIps = append(newIps, ip)
		}
	}
	rn.serfJoinedIps = newMap

	if len(newIps) == 0 {
		return nil
	}

	slog.InfoContext(ctx, "trying to join serf cluster", slog.Any("ipsCount", len(newIps)))
	n, _ := rn.serf.Join(newIps, false)
	slog.InfoContext(ctx, "finished joining", slog.Any("ipsCount", len(newIps)), slog.Any("joinedCount", n))

	return nil
}

func (rn *RunInfra) writeSerfMembersList(ctx context.Context) error {
	var sm types.SerfMembers
	for _, m := range rn.serf.Memberlist().Members() {
		sm.Members = append(sm.Members, types.SerfNode{
			Name: m.Name,
			Addr: m.Addr.String(),
		})
	}

	b, err := json.Marshal(sm)
	if err != nil {
		return err
	}

	if bytes.Equal(b, rn.serfMembersFileContent) {
		return nil
	}
	rn.serfMembersFileContent = b

	err = os.WriteFile(types.SerfMembersFile+".tmp", b, 0600)
	if err != nil {
		return err
	}
	err = os.Rename(types.SerfMembersFile+".tmp", types.SerfMembersFile)
	if err != nil {
		return err
	}
	return nil
}
