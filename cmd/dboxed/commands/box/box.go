package box

type BoxCommands struct {
	Create CreateCmd `cmd:"" help:"Create a box"`
	Get    GetCmd    `cmd:"" help:"Get a box"`
	List   ListCmd   `cmd:"" help:"List boxes" aliases:"ls"`
	Delete DeleteCmd `cmd:"" help:"Delete a box" aliases:"rm,delete"`
}
