//go:build linux

package runner

type RunnerCommands struct {
	Run     RunCmd     `cmd:"" help:"Download, unpack and run a box"`
	Systemd SystemdCmd `cmd:"" help:"Sub commands to control dboxed systemd integration"`
	Runc    RuncCmd    `cmd:"" help:"Run runc for a box"`

	RunBoxInSandbox RunBoxInSandbox `cmd:"" hidden:""`
}
