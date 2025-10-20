package compose

type ComposeCmd struct {
	Create CreateCmd `cmd:"" help:"Create a compose project"`
	Remove RemoveCmd `cmd:"" help:"Remove a compose project" aliases:"rm,delete"`
	List   ListCmd   `cmd:"" help:"List compose projects" aliases:"ls"`
	Update UpdateCmd `cmd:"" help:"Update a compose project"`
}
