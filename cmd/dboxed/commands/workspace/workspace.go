package workspace

type WorkspaceCommands struct {
	Create CreateCmd `cmd:"" help:"Create a workspace"`
	Delete DeleteCmd `cmd:"" help:"Delete a workspace" aliases:"rm,delete"`
	List   ListCmd   `cmd:"" help:"List workspaces" aliases:"ls"`
	Select SelectCmd `cmd:"" help:"Select a workspace"`
}
