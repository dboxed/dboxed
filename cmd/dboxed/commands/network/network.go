package network

type NetworkCommands struct {
	Create CreateCmd `cmd:"" help:"Create a network"`
	List   ListCmd   `cmd:"" help:"List networks" aliases:"ls"`
	Update UpdateCmd `cmd:"" help:"Update a network"`
	Delete DeleteCmd `cmd:"" help:"Delete a network" aliases:"rm,delete"`
}
