package volume

type VolumeCommands struct {
	Create            CreateCmd            `cmd:"" help:"Create a volume"`
	Delete            DeleteCmd            `cmd:"" help:"Delete a volume" aliases:"rm,delete"`
	List              ListCmd              `cmd:"" help:"List volumes" aliases:"ls"`
	ForceReleaseMount ForceReleaseMountCmd `cmd:"" help:"Force release a volume mount"`
}
