package spec

type SpecCommands struct {
	Create CreateCmd `cmd:"" help:"Create a dboxed spec"`
	Update UpdateCmd `cmd:"" help:"Update a dboxed spec"`
	List   ListCmd   `cmd:"" help:"List dboxed specs" aliases:"ls"`
	Delete DeleteCmd `cmd:"" help:"Delete a dboxed spec" aliases:"rm,delete"`
}
