package volume

type VolumeCommands struct {
	Create CreateCmd `cmd:"" help:"Create a volume"`
	Delete DeleteCmd `cmd:"" help:"Delete a volume"`
	List   ListCmd   `cmd:"" help:"List volumes"`

	Debug DebugCmd `cmd:"" help:"Debug commands"`

	OnlyLinuxCmds
}
