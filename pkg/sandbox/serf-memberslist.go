package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/koobox/unboxed/pkg/types"
	"github.com/koobox/unboxed/pkg/util"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (rn *Sandbox) runSerfStaticHosts(ctx context.Context) {
	for {
		err := rn.runSerfStaticHostsOnce(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "error in runSerfStaticHostsOnce", slog.Any("error", err))
		}

		if !util.SleepWithContext(ctx, 5*time.Second) {
			return
		}
	}
}

func (rn *Sandbox) runSerfStaticHostsOnce(ctx context.Context) error {
	b, err := os.ReadFile(filepath.Join(rn.getContainerRoot("_infra"), types.SerfMembersFile))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if bytes.Equal(b, rn.staticHostsMapBytes) {
		return nil
	}
	rn.staticHostsMapBytes = b

	var serfMembers types.SerfMembers
	err = json.Unmarshal(b, &serfMembers)
	if err != nil {
		return err
	}

	m := map[string]string{}
	for _, sm := range serfMembers.Members {
		fqdn := fmt.Sprintf("%s.%s", sm.Name, rn.BoxSpec.NetworkDomain)
		if !strings.HasSuffix(fqdn, ".") {
			fqdn += "."
		}
		m[fqdn] = sm.Addr
	}

	slog.InfoContext(ctx, "new static hosts map", slog.Any("staticHostsMap", m))

	rn.DnsProxy.SetStaticHostsMap(m)

	return nil
}
