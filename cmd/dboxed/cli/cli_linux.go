package cli

type cliOnlyLinux struct {
	Run     commands.RunCmd     `cmd:"" help:"Download, unpack and run a box"`
	Systemd commands.SystemdCmd `cmd:"" help:"Sub commands to control dboxed systemd integration"`
	Runc    commands.RuncCmd    `cmd:"" help:"Run runc for a box"`

	RunBoxInSandbox commands.RunBoxInSandbox `cmd:"" hidden:""`
}
