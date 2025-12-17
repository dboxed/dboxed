package git_credentials

type GitCredentialsCommands struct {
	Create CreateCmd `cmd:"" help:"Create git credentials"`
	Update UpdateCmd `cmd:"" help:"Update git credentials"`
	List   ListCmd   `cmd:"" help:"List git credentials" aliases:"ls"`
	Delete DeleteCmd `cmd:"" help:"Delete git credentials" aliases:"rm,delete"`
}
