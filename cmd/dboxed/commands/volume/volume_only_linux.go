package commands

type VolumeOnlyLinuxCmds struct {
	Lock    VolumeLockCmd    `cmd:"" help:"Lock a volume and prepare local image"`
	Release VolumeReleaseCmd `cmd:"" help:"Release a volume. A final backup will performed before actually releasing the volume"`
	Serve   VolumeServeCmd   `cmd:"" help:"Mount and sync a volume"`

	CleanupLoopDevs VolumeCleanupLoopDevs `cmd:"" help:"Find and clean up orphan loop devs and volumes that were created in garbage collected mount namespaces"`
}
