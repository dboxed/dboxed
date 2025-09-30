package cli

import (
	"github.com/dboxed/dboxed/cmd/dboxed/commands/systemd"
)

type cliOnlyLinux struct {
	Systemd systemd.SystemdCmd `cmd:"" help:"Sub commands to control dboxed systemd integration"`
}
