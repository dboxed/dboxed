package git_spec

type GitSpecCommands struct {
	Create CreateCmd `cmd:"" help:"Create a git spec"`
	Update UpdateCmd `cmd:"" help:"Update a git spec"`
	List   ListCmd   `cmd:"" help:"List git specs" aliases:"ls"`
	Delete DeleteCmd `cmd:"" help:"Delete a git spec" aliases:"rm,delete"`
}
