package run_infra_sandbox

import (
	"context"
	"fmt"
	"github.com/koobox/unboxed/pkg/util"
	"log/slog"
	"net"
	"reflect"
	"strings"
	"time"
)

func (rn *RunInfraSandbox) runNetbirdStatusToDns(ctx context.Context) {
	for {
		err := rn.runNetbirdStatusToDnsOnce(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "error in runNetbirdStatusToDnsOnce", slog.Any("error", err))
		}

		if !util.SleepWithContext(ctx, 5*time.Second) {
			return
		}
	}
}

func (rn *RunInfraSandbox) runNetbirdStatusToDnsOnce(ctx context.Context) error {
	s, err := rn.runNetbirdStatus(ctx)
	if err != nil {
		return err
	}

	ip, _, err := net.ParseCIDR(s.NetbirdIp)
	if err != nil {
		return err
	}

	localFqdn := fmt.Sprintf("%s.%s.", s.Name, rn.conf.BoxSpec.NetworkDomain)
	m := map[string]string{
		localFqdn: ip.String(),
	}
	for _, p := range s.Peers.Details {
		if p.Status != "Connected" {
			continue
		}

		fqdn := fmt.Sprintf("%s.%s", p.Name, rn.conf.BoxSpec.NetworkDomain)
		if !strings.HasSuffix(fqdn, ".") {
			fqdn += "."
		}
		m[fqdn] = p.NetbirdIp
	}

	if !reflect.DeepEqual(m, rn.olsStaticHostsMap) {
		slog.InfoContext(ctx, "new static hosts map", slog.Any("staticHostsMap", m))
		rn.dnsProxy.SetStaticHostsMap(m)
		rn.olsStaticHostsMap = m
	}

	return nil
}
