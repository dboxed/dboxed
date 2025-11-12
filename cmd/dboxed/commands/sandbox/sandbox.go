//go:build linux

package sandbox

import (
	"github.com/dboxed/dboxed/cmd/dboxed/commands/sandbox/service"
)

type SandboxCommands struct {
	Run    RunCmd    `cmd:"" help:"Run a box inside a sandbox"`
	List   ListCmd   `cmd:"" help:"List sandboxes" aliases:"ls"`
	Stop   StopCmd   `cmd:"" help:"Stop a sandbox"`
	Remove RemoveCmd `cmd:"" help:"Remove a sandbox" aliases:"rm,delete"`
	Exec   ExecCmd   `cmd:"" help:"execute new process inside the sandbox"`

	Service service.ServiceCmd `cmd:"" help:"Manage sandboxes as services"`

	RunInSandbox RunInSandbox `cmd:"" hidden:""`
	NetnsHolder  NetnsHolder  `cmd:"" hidden:""`
}
