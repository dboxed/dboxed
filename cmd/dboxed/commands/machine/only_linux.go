//go:build linux

package machine

type onlyLinux struct {
	Run RunCmd `cmd:"" help:"Run a machine and all its boxes" group:"run"`
}
