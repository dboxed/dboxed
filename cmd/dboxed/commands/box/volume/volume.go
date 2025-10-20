package volume

type VolumeCmd struct {
	Attach AttachCmd `cmd:"" help:"Attach a volume to a box"`
	Detach DetachCmd `cmd:"" help:"Detach a volume from a box"`
	Update UpdateCmd `cmd:"" help:"Update volume attachment settings"`
	List   ListCmd   `cmd:"" help:"List attached volumes" aliases:"ls"`
}
