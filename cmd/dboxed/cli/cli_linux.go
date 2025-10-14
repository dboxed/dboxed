package cli

import (
	"github.com/dboxed/dboxed/cmd/dboxed/commands/sandbox/service"
)

type cliOnlyLinux struct {
	Systemd service.ServiceCmd `cmd:"" help:"Sub commands to control dboxed systemd integration"`
}
