//go:build linux

package machine

import "github.com/dboxed/dboxed/cmd/dboxed/commands/machine/service"

type onlyLinux struct {
	Run RunCmd `cmd:"" help:"Run a machine and all its boxes" group:"run"`

	Service service.ServiceCmd `cmd:"" help:"Install and manage machine service"`
}
