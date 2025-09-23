package server

type ServerCommands struct {
	Run RunCmd `cmd:"" help:"run one or more server components"`
}
