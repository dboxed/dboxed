//go:build linux

package systemd

type SystemdCmd struct {
	Install SystemdInstallCmd `cmd:"" help:"Install dboxed as a systemd service"`
}
