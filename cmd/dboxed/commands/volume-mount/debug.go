package volume_mount

type DebugCmd struct {
	ResticRestServer ResticRestServerCmd `cmd:"" help:"Run a restic rest server"`
}
