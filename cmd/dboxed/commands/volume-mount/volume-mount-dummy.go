//go:build !linux

package volume_mount

type VolumeMountCommands struct {
	Debug DebugCmd `cmd:"" help:"Debug commands"`
}
