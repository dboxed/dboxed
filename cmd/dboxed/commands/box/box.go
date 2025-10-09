package box

type BoxCommands struct {
	Create CreateCmd `cmd:"" help:"Create a box"`
	Get    GetCmd    `cmd:"" help:"Get a box"`
	List   ListCmd   `cmd:"" help:"List boxes"`
	Delete DeleteCmd `cmd:"" help:"Delete a box"`
}
