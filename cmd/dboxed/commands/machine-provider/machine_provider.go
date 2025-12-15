package machine_provider

type MachineProviderCommands struct {
	Create CreateCmd `cmd:"" help:"Create a machine provider"`
	Update UpdateCmd `cmd:"" help:"Update a machine provider"`
	List   ListCmd   `cmd:"" help:"List machine providers" aliases:"ls"`
	Delete DeleteCmd `cmd:"" help:"Delete a machine provider" aliases:"rm,delete"`
}
