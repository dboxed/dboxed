package volume_provider

type VolumeProviderCommands struct {
	Create CreateCmd `cmd:"" help:"Create a provider"`
	Update UpdateCmd `cmd:"" help:"Update a provider"`
	Delete DeleteCmd `cmd:"" help:"Delete a provider"`
	List   ListCmd   `cmd:"" help:"List providers"`
}
