//go:build linux

package service

type ServiceCmd struct {
	Install   InstallCmd   `cmd:"" help:"Install dboxed as a service"`
	Uninstall UninstallCmd `cmd:"" help:"Uninstall dboxed as a service"`
	Start     StartCmd     `cmd:"" help:"Start dboxed as a service"`
	Stop      StopCmd      `cmd:"" help:"Stop dboxed as a service"`
}
