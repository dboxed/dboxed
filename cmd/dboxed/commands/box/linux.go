//go:build linux

package box

type onlyLinux struct {
	Run  RunCmd  `cmd:"" help:"Run a box"`
	Runc RuncCmd `cmd:"" help:"Run runc for a box"`

	RunInSandbox RunInSandbox `cmd:"" hidden:""`
}
