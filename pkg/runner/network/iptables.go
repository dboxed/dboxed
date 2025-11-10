//go:build linux

package network

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"syscall"
	"text/template"
)

// we can't rely on the host iptables binary to be present or functioning, so we enter the sandbox
// via unshare+chroot, mount some necessary stuff and then perform the iptables inside the sandbox.
// this will NOT change the network NS, so we actually modify the host iptables
const baseScript = `
set -e

mount -t proc none /proc
mount -t sysfs none /sys
`

type Iptables struct {
	InfraContainerRoot string
	NamesAndIps        NamesAndIps
}

func (n *Iptables) runIptablesScript(ctx context.Context, script string) error {
	script2 := baseScript
	script2 += fmt.Sprintf("export NAME_BASE='%s'", n.NamesAndIps.Base) + "\n"
	script2 += script

	slog.DebugContext(ctx, "running iptables script:\n"+script2+"\n")

	cmd := exec.CommandContext(ctx, "chroot", n.InfraContainerRoot, "/bin/sh", "-c", script2)
	cmd.Env = []string{
		"PATH=/sbin:/bin:/usr/sbin:/usr/bin",
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Unshareflags: syscall.CLONE_NEWNS}
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (n *Iptables) runPurgeOldRules(ctx context.Context) error {
	slog.InfoContext(ctx, "purging old dboxed iptables rules")
	script := `
OLD_RULES="$(iptables-save)"
if echo "$OLD_RULES" | grep "\--comment ${NAME_BASE}" > /dev/null; then
  echo "$OLD_RULES" | grep -v "\--comment ${NAME_BASE}" | iptables-restore
fi
if echo "$OLD_RULES" | grep "^:${NAME_BASE}-pf-1" > /dev/null; then
  iptables -t nat -F ${NAME_BASE}-pf-1
  iptables -t nat -X ${NAME_BASE}-pf-1
fi
if echo "$OLD_RULES" | grep "^:${NAME_BASE}-pf-2" > /dev/null; then
  iptables -t nat -F ${NAME_BASE}-pf-2
  iptables -t nat -X ${NAME_BASE}-pf-2
fi
`
	return n.runIptablesScript(ctx, script)
}

func (n *Iptables) setupIptables(ctx context.Context) error {
	err := n.runPurgeOldRules(ctx)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "setting up iptables rules")

	t, err := template.New("").Parse(`
COMMENT="-m comment --comment ${NAME_BASE}"

iptables $COMMENT -A FORWARD -o {{ .HostInterface }} -j ACCEPT
iptables $COMMENT -A FORWARD -i {{ .HostInterface }} -j ACCEPT
iptables $COMMENT -t nat -A POSTROUTING -s {{ .HostAddr }} -j MASQUERADE

iptables -t nat -N ${NAME_BASE}-pf-1
iptables -t nat -N ${NAME_BASE}-pf-2
`)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(nil)
	err = t.Execute(buf, map[string]string{
		"HostInterface": n.NamesAndIps.VethNameHost,
		"HostAddr":      n.NamesAndIps.HostAddr.IPNet.String(),
	})
	if err != nil {
		return err
	}

	return n.runIptablesScript(ctx, buf.String())
}
