//go:build linux

package sandbox

import (
	"fmt"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/util"
)

func (rn *Sandbox) writeDnsProxyResolvConf() error {
	resolveConf := fmt.Sprintf(`# This is the dboxed dns proxy, which listens inside the sandboxed network namespace
# and forwards requests to the host's resolv.conf nameservers
nameserver %s
search .
`, rn.network.Config.DnsProxyIP)

	err := util.AtomicWriteFile(filepath.Join(rn.GetSandboxRoot(), "/etc/resolv.conf"), []byte(resolveConf), 0644)
	if err != nil {
		return err
	}
	return nil
}
