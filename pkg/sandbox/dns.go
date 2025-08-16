package sandbox

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"time"

	dns_proxy "github.com/dboxed/dboxed/pkg/dns-proxy"
	"github.com/dboxed/dboxed/pkg/types"
	"github.com/dboxed/dboxed/pkg/util"
)

func (rn *Sandbox) startDnsProxy(ctx context.Context) error {
	rn.dnsProxy = &dns_proxy.DnsProxy{
		ListenNamespace: rn.network.NetworkNamespace,
		QueryNamespace:  rn.network.HostNetworkNamespace,
		ListenIP:        net.ParseIP(rn.network.Config.DnsProxyIP),
		HostFsPath:      "/",
	}

	err := rn.dnsProxy.Start(ctx)
	if err != nil {
		return err
	}

	go rn.runReadDnsMap(ctx)

	return nil
}

func (rn *Sandbox) runReadDnsMap(ctx context.Context) {
	util.LoopWithPrintErr(ctx, "runReadDnsMapOnce", 5*time.Second, func() error {
		return rn.runReadDnsMapOnce(ctx)
	})
}

func (rn *Sandbox) runReadDnsMapOnce(ctx context.Context) error {
	b, err := os.ReadFile(filepath.Join(rn.getInfraRoot(), types.DnsMapFile))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	h := util.Sha256Sum(b)
	if h == rn.oldDnsMapHash {
		return nil
	}

	var m map[string]string
	err = json.Unmarshal(b, &m)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "updating dns proxy static hosts map", slog.Any("dnsMap", m))

	rn.dnsProxy.SetStaticHostsMap(m)

	rn.oldDnsMapHash = h

	return nil
}
