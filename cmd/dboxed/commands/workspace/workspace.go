package workspace

type WorkspaceCommands struct {
	Create CreateCmd `cmd:"" help:"Create a workspace"`
	Delete DeleteCmd `cmd:"" help:"Delete a workspace"`
	List   ListCmd   `cmd:"" help:"List workspaces"`
	Select SelectCmd `cmd:"" help:"Select a workspace"`
}
