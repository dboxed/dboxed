//go:build linux

package volume_mount

import "github.com/dboxed/dboxed/cmd/dboxed/commands/volume-mount/service"

type VolumeMountCommands struct {
	Create  CreateCmd  `cmd:"" help:"Create volume mount"`
	Release ReleaseCmd `cmd:"" help:"Release a volume. A final backup will performed before actually releasing the volume"`
	List    ListCmd    `cmd:"" help:"List a volume mounts" aliases:"ls"`
	Mount   MountCmd   `cmd:"" help:"Mount a volume"`
	Serve   ServeCmd   `cmd:"" help:"Mount and sync a volume"`

	Service service.ServiceCmd `cmd:"" help:"Manage volume mount services"`

	CleanupLoopDevs CleanupLoopDevs `cmd:"" help:"Find and clean up orphan loop devs and volumes that were created in garbage collected mount namespaces"`

	Debug DebugCmd `cmd:"" help:"Debug commands"`
}
