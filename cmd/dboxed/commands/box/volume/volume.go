package volume

type VolumeCmd struct {
	Attach AttachCmd `cmd:"" help:"Attach a volume to a box"`
	Detach DetachCmd `cmd:"" help:"Detach a volume from a box"`
	List   ListCmd   `cmd:"" help:"List attached volumes" aliases:"ls"`
}
