package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
)

func (rn *Sandbox) writeResolvConf(root string, dnsProxyIp string) error {
	resolveConf := fmt.Sprintf(`# This is the dboxed dns proxy, which listens inside the sandboxed network namespace
# and forwards requests to the host's resolv.conf nameservers
nameserver %s
search .
`, dnsProxyIp)

	err := os.WriteFile(filepath.Join(root, "etc/resolv.conf"), []byte(resolveConf), 0666)
	if err != nil {
		return err
	}
	return nil
}
