package machine

type MachineCommands struct {
	Create CreateCmd `cmd:"" help:"Create a machine"`
	List   ListCmd   `cmd:"" help:"List machines" aliases:"ls"`
	Delete DeleteCmd `cmd:"" help:"Delete a machine" aliases:"rm,delete"`

	AddBox    AddBoxCmd    `cmd:"" help:"Add a box to a machine" group:"box"`
	RemoveBox RemoveBoxCmd `cmd:"" help:"Remove a box from a machine" group:"box" aliases:"rm-box"`
	ListBoxes ListBoxesCmd `cmd:"" help:"List boxes for a machine" aliases:"ls-boxes" group:"box"`

	onlyLinux
}
