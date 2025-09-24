package auth

type AuthCommands struct {
	Login LoginCmd `cmd:"" help:"login to the dboxed api"`

	Token TokenCmd `cmd:"" help:"manage api tokens"`
}
