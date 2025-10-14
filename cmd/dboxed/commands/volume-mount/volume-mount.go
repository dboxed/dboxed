//go:build linux

package volume_mount

type VolumeMountCommands struct {
	Lock    LockCmd    `cmd:"" help:"Lock a volume and prepare local image"`
	Release ReleaseCmd `cmd:"" help:"Release a volume. A final backup will performed before actually releasing the volume"`
	Serve   ServeCmd   `cmd:"" help:"Mount and sync a volume"`

	CleanupLoopDevs CleanupLoopDevs `cmd:"" help:"Find and clean up orphan loop devs and volumes that were created in garbage collected mount namespaces"`

	Debug DebugCmd `cmd:"" help:"Debug commands"`
}
