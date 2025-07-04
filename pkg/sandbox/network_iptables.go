package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"github.com/koobox/unboxed/pkg/util"
	"github.com/vishvananda/netlink"
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

export PATH=/bin:/sbin

mount -t proc none /proc
mount -t sysfs none /sys

echo "checking nft mode, please ignore errors that you might see in the next lines"
if iptables-nft -L > /dev/null; then
  echo "using iptables-nft"
  IPTABLES=iptables-nft
  IPTABLES_SAVE=iptables-nft-save
  IPTABLES_RESTORE=iptables-nft-restore
else
  echo "using iptables-legacy"
  IPTABLES=iptables-legacy
  IPTABLES_SAVE=iptables-legacy-save
  IPTABLES_RESTORE=iptables-legacy-restore
fi
`

func (rn *Sandbox) runIptablesScript(ctx context.Context, script string) error {
	script2 := baseScript + "\n"
	script2 += fmt.Sprintf("export NAME_PREFIX='%s'", rn.buildIptablesNamePrefix()) + "\n"
	script2 += script

	slog.InfoContext(ctx, "running iptables script:\n"+script2+"\n")

	cmd := exec.CommandContext(ctx, "chroot", rn.getContainerRoot("_infra"), "/bin/sh", "-c", script2)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Unshareflags: syscall.CLONE_NEWNS}
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (rn *Sandbox) runPurgeOldRules(ctx context.Context) error {
	script := `
OLD_RULES="$($IPTABLES_SAVE)"
if echo "$OLD_RULES" | grep "\--comment ${NAME_PREFIX}" > /dev/null; then
  echo "purging old unboxed iptables rules"
  echo "$OLD_RULES" | grep -v "\--comment ${NAME_PREFIX}" | $IPTABLES_RESTORE
fi
if echo "$OLD_RULES" | grep "^:${NAME_PREFIX}-pf-1" > /dev/null; then
  $IPTABLES -t nat -F ${NAME_PREFIX}-pf-1
  $IPTABLES -t nat -X ${NAME_PREFIX}-pf-1
fi
if echo "$OLD_RULES" | grep "^:${NAME_PREFIX}-pf-2" > /dev/null; then
  $IPTABLES -t nat -F ${NAME_PREFIX}-pf-2
  $IPTABLES -t nat -X ${NAME_PREFIX}-pf-2
fi
`
	return rn.runIptablesScript(ctx, script)
}

func (rn *Sandbox) setupIptables(ctx context.Context, hostAddr netlink.Addr) error {
	log := slog.With()
	log.InfoContext(ctx, "setting up iptables rules")

	err := rn.runPurgeOldRules(ctx)
	if err != nil {
		return err
	}

	t, err := template.New("").Parse(`
COMMENT="-m comment --comment ${NAME_PREFIX}"

$IPTABLES $COMMENT -A FORWARD -o {{ .HostInterface }} -j ACCEPT
$IPTABLES $COMMENT -A FORWARD -i {{ .HostInterface }} -j ACCEPT
$IPTABLES $COMMENT -t nat -A POSTROUTING -s {{ .HostAddr }} -j MASQUERADE

$IPTABLES -t nat -N ${NAME_PREFIX}-pf-1
$IPTABLES -t nat -N ${NAME_PREFIX}-pf-2
`)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(nil)
	err = t.Execute(buf, map[string]string{
		"HostInterface": rn.vethNameHost,
		"HostAddr":      hostAddr.IPNet.String(),
	})
	if err != nil {
		return err
	}

	return rn.runIptablesScript(ctx, buf.String())
}

func (rn *Sandbox) buildIptablesNamePrefix() string {
	h := util.Sha256Sum([]byte(rn.SandboxName))
	h = h[:6]

	return fmt.Sprintf("ub-%s", h)
}
