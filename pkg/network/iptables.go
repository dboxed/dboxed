package network

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"text/template"
)

type Iptables struct {
	NamesAndIps NamesAndIps
}

func (n *Iptables) runIptablesScript(ctx context.Context, script string) error {
	script2 := "set -e\n"
	script2 += fmt.Sprintf("export NAME_BASE='%s'", n.NamesAndIps.Base) + "\n"
	script2 += script

	slog.InfoContext(ctx, "running iptables script:\n"+script2+"\n")

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", script2)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (n *Iptables) runPurgeOldRules(ctx context.Context) error {
	script := `
OLD_RULES="$(iptables-save)"
if echo "$OLD_RULES" | grep "\--comment ${NAME_BASE}" > /dev/null; then
  echo "purging old unboxed iptables rules"
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
	log := slog.With()
	log.InfoContext(ctx, "setting up iptables rules")

	err := n.runPurgeOldRules(ctx)
	if err != nil {
		return err
	}

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
