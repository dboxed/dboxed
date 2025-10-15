package service

type ServiceCmd struct {
	Install   InstallCmd   `cmd:"" help:"Install a dboxed-volume system service"`
	Uninstall UninstallCmd `cmd:"" help:"Uninstall a dboxed-volume system service"`
	Start     StartCmd     `cmd:"" help:"Start a dboxed-volume system service"`
	Stop      StopCmd      `cmd:"" help:"Stop a dboxed-volume system service"`
}
