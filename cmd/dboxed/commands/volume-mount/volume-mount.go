//go:build linux

package volume_mount

type VolumeMountCommands struct {
	Release     ReleaseCmd     `cmd:"" help:"Release a volume. A final backup will performed before actually releasing the volume"`
	ForceRemove ForceRemoveCmd `cmd:"" help:"Force remove a volume mount without backup (emergency cleanup)"`
	List        ListCmd        `cmd:"" help:"List volume mounts" aliases:"ls"`
	Create      CreateCmd      `cmd:"" help:"Create volume mount"`
	Mount       MountCmd       `cmd:"" help:"Mount a volume"`
	Serve       ServeCmd       `cmd:"" help:"Mount and sync a volume"`

	CleanupLoopDevs CleanupLoopDevs `cmd:"" help:"Find and clean up orphan loop devs and volumes that were created in garbage collected mount namespaces"`

	Debug DebugCmd `cmd:"" help:"Debug commands"`
}
